// classifyFilesLib.go
package qprob

/* classifyFile.go  - Train from 1 CSV File and classify using
using a second test file.  Output statistics on classification
from algorithm versus actual classification */

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"qutil"
	//"strconv"
	//"encoding/json"
	"io/ioutil"
	s "strings"
)

const DefaultCSVMaxLoadSize = OneGig

//func check(msg string, e error) {
//	if e != nil {
//		fmt.Println("ERROR:")
//		fmt.Println(e)
//		panic(e)
//	}
//}

func printLn(f *os.File, txt string) {
	_, err1 := f.WriteString(txt)
	check("err in printLn ", err1)

	_, err2 := f.WriteString("\n")
	check("err in printLn ", err2)
}

func ProcessRowsRows(fier *Classifier, req *ClassifyRequest, rows [][]float32, inName string, outBaseName string, asTest bool) {
	fmt.Println("\nfinished build Now Try to classify")
	detRows, sumRows := fier.ClassifyRows(rows, fier.ColDef)
	classCol := fier.ClassCol
	//
	fmt.Printf("num deRows=%v\n", len(detRows))
	fmt.Printf("num sumRows.Precis=%v\n", sumRows.Precis)

	if req.WriteJSON {
		jsonsumstr := sumRows.ToJSON()
		outFileName := s.Replace(outBaseName, ".csv", ".out.sum.json", -1)
		outFileName = s.Replace(outFileName, ".out.out", ".out", -1)
		fmt.Printf("write JSON sum rows to %s\n", outFileName)
		if req.DetToStdOut {
			fmt.Printf("sumRows asJSON=%s\n", jsonsumstr)
		}
		ioutil.WriteFile(outFileName, jsonsumstr, 0644)

		// Add class probability output
		// add detailed probability output
	}

	if req.DoPreAnalyze {
		// Pre-analyze each column to try and find the sweet spot
		// for precision and recall as number of buckets.
		origTrainRows := fier.GetTrainRowsAsArr(OneGig)
		testRows := origTrainRows
		trainRows := origTrainRows
		if req.AnalSplitType == 1 {
			// pull test records from the body of test data
			// best for normal sets.
			oneEvery := int(float32(len(origTrainRows)) / (float32(len(origTrainRows)) * req.AnalTestPort))
			fmt.Printf("Analyze SplitOneEvery=%v  portSet=%v\n", oneEvery, req.AnalTestPort)
			trainRows, testRows = qutil.SplitFloatArrOneEvery(origTrainRows, 1, oneEvery)
		} else {
			// pull records from end of test data.  Best for
			// time series when predicting on records near the end
			fmt.Printf("Analyze splitEnd PortSet=%v", req.AnalTestPort)
			trainRows, testRows = qutil.SplitFloatArrTail(origTrainRows, req.AnalTestPort)
		}
		// Have to retrain based on the newly split data
		fmt.Printf("Analyze #TrainRow=%v #TestRow=%v\n", len(trainRows), len(testRows))
		fier.Retrain(trainRows)
		anaRes := fier.TestIndividualColumnsNB(AnalNoClassSpecified, -1.0, trainRows, testRows)
		fmt.Printf("L68: anaRes=%v\n", anaRes)

		// Have to re-run with the new configuration
		// of column settings
		fier.Retrain(origTrainRows)
		detRows, sumRows = fier.ClassifyRows(rows, fier.ColDef)
	}

	// Convert the summary Rows into printable Output to display
	if req.WriteCSV {
		var sbb bytes.Buffer
		sb := &sbb
		if asTest {
			sumRows.AsStrToBuffTest(sb)
		} else {
			sumRows.AsStrToBuffClass(sb)
		}
		outFileName := s.Replace(outBaseName, ".csv", ".out.sum.csv", -1)
		outFileName = s.Replace(outFileName, ".out.out", ".out", -1)
		fmt.Printf("write CSV sum rows to %s\n", outFileName)
		ioutil.WriteFile(outFileName, sb.Bytes(), 0644)
		if req.DetToStdOut {
			// TODO: Need to re-read actual string values
			// from the input CSV rather than the parsed value
			// otherwise those that did not have valid numbers
			// do not work.
			fmt.Printf("As Disp Str\n%s\n", sb.String())
			failCnt := sumRows.TotCnt - sumRows.SucCnt
			failP := 1.0 - sumRows.Precis
			fmt.Printf("numRow=%v  sucCnt=%v precis=%v failCnt=%v failPort=%v",
				sumRows.TotCnt, sumRows.SucCnt, sumRows.Precis, failCnt, failP)
		}

		// add class probability output

		// add detailed probability output
	}

	if req.WriteFullCSV {
		// copy the input CSV to output CSV
		// changing only the class column to
		// the predicted class
		var sbb bytes.Buffer
		sb := &sbb
		fiIn, err := os.Open(inName)
		check("opening file", err)
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(fiIn)
		defer fiIn.Close()

		// Copy of header to both files
		scanner.Scan() // skip headers
		headTxt := s.TrimSpace(scanner.Text())
		fmt.Fprintln(sb, headTxt)
		rowndx := 0
		for scanner.Scan() {
			txt := s.TrimSpace(scanner.Text())

			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}

			flds := s.Split(txt, ",")
			flds[classCol] = fmt.Sprintf("%v", sumRows.Rows[rowndx].BestClass)
			rowstr := s.Join(flds, ",")
			fmt.Fprintln(sb, rowstr)
			rowndx += 1
		} // for row

		outFileName := s.Replace(outBaseName, ".csv", ".out.csv", -1)
		outFileName = s.Replace(outFileName, ".out.out", ".out", -1)
		fmt.Printf("write CSV detail rows to %s\n", outFileName)
		ioutil.WriteFile(outFileName, sb.Bytes(), 0644)
		if req.DetToStdOut {
			fmt.Println("Original CSV with Class Column changed to predicted class")
			fmt.Println(sb.String())
		}
	}
	fmt.Printf("\nSummary By Class\n")
	clasSum := fier.MakeByClassStats(sumRows, rows)
	fier.PrintTrainClassProb()
	fmt.Printf("\nRESULTS FOR TEST DATA\n  Num Test Rows=%v\n", len(rows))
	fier.PrintResultsByClass(clasSum)
	// TODO: Print this out nicely
}

/*
  Uses the training file to train the classifier
  then reads Reads lines out of test CSV file
  having the classifier classify each line
  and report how accurate the classifier is.
  reports both precision and recall by class.

  NOTE: Moved to processingRows for Test because
   I wanted that methd for use with the optimizer
   not quite as scalable as the line by line but
   we also needed different options for output.
*/
func ClassifyTestFiles(req *ClassifyRequest) {

	fmt.Printf("\rClassifyFiles\r  trainFiName=%s\r  testFiName=%s\r  maxNumBuck=%v\n",
		req.TrainInFi, req.TestInFi, req.MaxNumBuck)

	fier := LoadClassifierTrainFile(req.TrainInFi, "test", req.MaxNumBuck)
	fier.Req = req

	fmt.Println("constructor complete")
	//fmt.Println(fier.String())

	// If we were processing a test file then
	// save the output to represent it's results
	if req.TestInFi != "" {
		fmt.Printf("processing test data %s", req.TestInFi)
		header, testRows := LoadCSVRows(req.TestInFi, DefaultCSVMaxLoadSize)
		req.Header = header
		fmt.Printf("Loaded %v rows\n", len(testRows))

		if req.DoOpt {
			fier.OptProcess(1, req.OptMaxTime, 3.0)
		}

		ProcessRowsRows(fier, req, testRows, req.TestInFi, req.TestOutFi, true)
	}

	// If we have a classification job then process it.
	if req.ClassInFi != "" {
		fmt.Printf("processing classify data %s", req.ClassInFi)
		header, rows := LoadCSVRows(req.ClassInFi, DefaultCSVMaxLoadSize)
		req.Header = header
		fmt.Printf("Loaded %v rows\n", len(rows))
		ProcessRowsRows(fier, req, rows, req.ClassInFi, req.ClassOutFi, false)
	}

}

/*
  Uses the training file to train the classifier
  then reads Reads lines out of test CSV file
  having the classifier classify each line
  and report how accurate the classifier is.
  reports both precision and recall by class.
*/
func ClassifyTestFilesLargeFile(req *ClassifyRequest) {
	var sbb bytes.Buffer
	sb := &sbb
	trainFiName := req.TrainInFi
	testFiName := req.TestInFi
	numBuck := req.MaxNumBuck

	fmt.Fprintf(sb, "\rClassifyFiles\r  trainFiName=%s\r  testFiName=%s\r  numBuck=%v\n",
		trainFiName, testFiName, numBuck)

	fier := LoadClassifierTrainFile(trainFiName, "test", numBuck)
	fmt.Fprintln(sb, "constructor complete")
	//fmt.Println(fier.String())

	fmt.Fprintln(sb, "\nfinished build Now Try to classify")
	fiIn, err := os.Open(testFiName)
	check("opening file", err)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(fiIn)
	defer fiIn.Close()

	// Copy of header to both files
	scanner.Scan() // skip headers
	//headTxt := s.TrimSpace(scanner.Text())

	fmt.Fprintln(sb, "row,predClass,predProb,actClass,status\n")
	// Copy the rows.
	sucessCnt := 0
	rowCnt := 0
	for scanner.Scan() {
		txt := s.TrimSpace(scanner.Text())

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		cres := fier.ClassRowStr(txt)
		flds := ParseStrAsArrInt32(txt)
		actualClass := int16(flds[fier.ClassCol])
		statTxt := "sucess"
		if actualClass != cres.BestClass {
			statTxt = "fail  "
		} else {
			sucessCnt += 1
		}
		// TODO: We want to track sucess by class

		fmt.Fprintf(sb, "%v,%v,%v,%v,%v\n",
			rowCnt, cres.BestClass, cres.BestProb, actualClass, statTxt)

		rowCnt += 1
	} // for row

	percCorr := (float32(sucessCnt) / float32(rowCnt)) * float32(100.0)
	percFail := 100.0 - percCorr

	fmt.Fprintf(sb, "tested %v rows  with %v correct sucess=%v%% fail=%v%%",
		rowCnt, sucessCnt, percCorr, percFail)
	sb.WriteTo(os.Stdout)
}

func printClassifyFilesHelp() {
	msg := `
    -train=finame      file containing training data
                       optional when model input is specified
  
    -test=finame       file containing data to use to test model
	                   file must exist when specified. 
                       optional when -class is specified. 
					
    -class=finame      name of file containing data to classify 
	                   must exist is specified.   Optional when
					   -test is specified.  By convention class
					   is set to -1 in input class files but the
					   system doese not care. 
	
	-classout=finame   name of file to write classify results to
	                   will be written in csv format.  If not
					   specified default name  will be name 
					   specified by -class with .csv
					   replaced with .out.csv.  By convention
					   all named output files should end with
					   .out.csv. 
	
	-testout=finame    Write test output CSV file name to this 
	                   file instead of the default output file.
					   By convention all output files should 
					   end with .out.csv
					
	-maxBuck=10        Maximum number of buckets the system is 
	                   allowed to use when computing quanta. This
					   can significantly increase RAM and under
					   some conditions it is desirable to limit
					   maxNum to prevent too constrained of
					   selection.  Defaults to 10
	
	-writeJSON=true    if present then write results to JSON files
	                   otherwise will only generate CSV.
					
  -writeCSV=true       Will write output in CSV form which will 
	                   require multiple files in some instances
					   or supress some explanatory information
					   defaults to true if not specified.
					
  -writeFullcsv=false  Write the original CSV with all columns
	                   the same except for the class column values
					   will be changed to the predicted class 
					   defaults to false.
					
  -writeDetails=true   Write files containing detailed 
	                   probability by row in addition to the
					   summary information this shows the
					   probability of each row belonging
					   to each class. file extensions will 
					   be .det added to path  name. 
					
  -detToStdOut=false   When true will print values saved in the
	                   generated files to stdout as things are
					   processed.  This consumes  considerable
					   time so turn of except when debugging. 
					   defaults to true. 
					
  -doOpt=false.        When set to true the optimizer is allowed
                       to run and use all the -opt* features 
					   but when false will not run optimizer at
					   all but it will still load existing optimizer
					   settings if the file exists.
  
  -optMaxTime=1        Max Time each cycle of Optimizer is allowed 
	                   to run when seeking to reach optMinPrec
					   Specified as integer num representing 
					   seconds between 1.0 and 1000.0 Defaults to 
					   1 if not specified. 
					
  -OptClassId=-999    Normall optimizer logic involves either
                       cycling time between all classes or
					   simply trying to maximize precision 
					   measured across all documents.  When
					   this number value is set to a valid classid
					   then all optimizer time will be invested
					   in improving performance for that class
					   which may happen at the expense of 
					   precision and recall for other classes 
					   defaults to -999
					
  -OptCycleClass=true  When true and when a specific class for 
                       optimization has not been supplied 
					   the optimizer will cycle through features
					   optimzing each one with the knowledge that
					   a change to help one may hurt others.  
					   When false and no class has been specified
					   using the -optClassId flag then optimizer
					   only looks at precision for entire set
					   of test records which means that the 
					   class with the most records in the training
					   set will get preferential treatment.
					   defaults to true  TBD 
					
					
  -optClear=false	   Clear existing optimizer settings 
                       to a weight of 1 and numBuck equal
					   to current setting for -NumBuck. 
					   This will cause any currently saved 
					   optimizer settings  to be cleared. 
					   This will delete the optimizer savings
					   file generated by prior optimizer runs.
					   TBD
					
  -optSave=true        Will cause the new optimizer settings
                       to be saved so they can automatically
					   be restored for the next run with the 
					   same input file.  This is saved in 
					   same directory with inputTraining file
					   with name changed from finame.csv to
					   finame.optset.txt
					   TBD
	
  -optMinRecall=0.01   When optimizing for specific features 
                       the optimizer normally only keeps a 
					   change if it increases precision 
					   or increases recall without reducing
					   precision.  The optimizer is allowed
					   to keep changes that increase recall
					   even when it hurts precision when current
					   recall is below this number.  
					   Defaults to 0.01 which means it will
					   agressively seek to increase recall whenever
					   recall is below 1%.
					
  -OptMaxPrec=0.95     When optimizing at the feature level the 
                       system normally seeks to always increase
					   precision but under some cases once we
					   have good enough precision it is better 
					   to increase recall.  Once the precision
					   for a class exceeds this number the 
					   optimizer will accept changes provided 
					   they increase recall and do not reduce
					   precision below this value.   Defaults
					   to 0.95 if not set.
  
  -optRandomize=false  If true will randomize all current 
                       optimizer settings.  This can 
					   help discover new paths but is really in place
					   to support genetic altorithms in the future.
					   defaults to false
	
	
A integer ID that maps to one class.  The Optimizer
	                   run compare the results for this class to try and
					   increase recall and precsion rather than improving
					   precision for the entire set.  This is used for
					   things like stocks where I really want to be more
					   accurate when picking bars that will increase in 
					   price and earn a profit.
					
	
					
  eg: classifyFiles -train=data/titanic.train.csv -test=data/titanic/test.csv -numBuck=6
	 will read training file to build classifer
	 will read test data file to classify each line
     will report results of test versus repored class
	 classifier will use numBuck buckets.
	 classified test records will be written to data/titanic/test.out.csv 
	 classified test summary will be written to data/titanic/test.out.meta.txt
	
	
  eg: classifyFiles -train=data/titanic.train.csv -class=data/toclassify.csv")
	 Will read training file data/titanic.train.csv  build the model
	 and will read file data/toclassify.csv as records to classify.
	 The resulting output will be written to data/toclassify.out.csv 
	 because no -classout was specified.  Numuck defaulted to 10 because
	 was not specified.
	
   It is legal to specify both -class and -test it will run the test 
   and then run the classify.

   It is legal to use a filter generated as a -test input file as input
   to the -class option.   
					
    -------------------
	---- FUTURE -------
    -------------------
	
					
	-readModel=true   If defined and set to true and when -model is set
	                   the exisitng model file will be read before the
					   training file.  Otherwise the model file will be
					   ignored and replaced with data from the training
					   phase.  Defaults to true if not specified
					
	-model             Name of file to read as the model from. It will
	                   be read if the model before the training data if
					   it exists.  If File does not exist it will be 
					   generated when the training has been completed
					   and after optimization passes.
					
	-runOpt=true       If true will run the optimizer.  If false then
	                   all optimizer settings are ignored. 	 Defaults
					   to false
				

					
    -optMinPrec=95     Will run optimizer cycles until a minimum 
	                   precision at 100% recall has been reached.
					   will stop when optimizer has ran optMaxCycles
					   if not set then no optimizer is ran. 
	
  
					
	
	-optOKBuck=true    If true then optimizer is allowed to 
	                   change numer of quant buckets.  Defaults
					   to true if not set.  Must be true if
					   -optOKWeight is false when he optimizer 
					   is set. 
					
	-optOKWeight       If True the optimizer is allowed to change
	                   feature weight.  Defaults to true if not set.
					   Must be true if -optOKBuck is false when
					   optimizer is ran. 
					
	
	-DoPreAnalyze      If true will pre-analyze data set attempting
	                   to find number of buckets for each column
					   that maximizes a combination of precision and
					   recall.  Defaults to false
					
    -AnalClassId       When specified it will analyze based on finding
	                   best settings for the specified class otherwise
					   will try to find best precision for the entire
					   set of records.  Defaults to not set
	
	-AnalSplitType     if 1 then split by pulling test records from
	                   body of training data.  If 2 then pull test
					   records from end of training data. Defaults
					   to 1
	
	-AnalSplitPort     The portion of training set to use as test data
	                   defaults to 0.15 if not set.
	
  -`
	fmt.Println(msg)
}

func checkClassifyFilesParms(msg string, abort bool) {
	if abort {
		fmt.Printf("ERROR: %s", msg)
		printClassifyFilesHelp()
		log.Fatal("Exit")
	}
}

func defCSVOutName(str string) string {
	ts := str
	ts = s.Replace(ts, ".csv", ".out.csv", -1)
	return ts
}

func defModelName(str string) string {
	ts := str
	ts = s.Replace(ts, "csv", ".model.csv", -1)
	return ts
}

func ParseClassifyFileCommandParms(args []string) *ClassifyRequest {
	aReq := MakeEmptyClassifyFilesRequest()
	parms := qutil.ParseCommandLine(args)
	fmt.Println(parms.String())

	aReq.TrainInFi = parms.Sval("train", "")
	aReq.TestInFi = parms.Sval("test", "")
	aReq.ClassInFi = parms.Sval("class", "")
	aReq.ModelFi = parms.Sval("model", aReq.TrainInFi)
	aReq.ClassOutFi = parms.Sval("classout", defCSVOutName(aReq.ClassInFi))
	aReq.TestOutFi = parms.Sval("testout", defCSVOutName(aReq.TestInFi))
	aReq.MaxNumBuck = int16(parms.Ival("maxbuck", 10))
	aReq.LoadModel = parms.Bval("loadModel", true)
	aReq.WriteJSON = parms.Bval("writejson", false)
	aReq.WriteCSV = parms.Bval("writeCSV", true)
	aReq.WriteDetails = parms.Bval("writedetails", false)
	aReq.WriteFullCSV = parms.Bval("writefullcsv", false)
	aReq.DetToStdOut = parms.Bval("dettostdout", true)
	aReq.DoOpt = parms.Bval("doopt", false)
	aReq.OptPreRandomize = parms.Bval("optrandomize", false)
	aReq.OptMaxTime = parms.F64val("optmaxtime", 2.0) * 1000.0
	aReq.OptClassId = int16(parms.Ival("optclassid", AnalNoClassSpecified))
	aReq.OptMinRecall = parms.Fval("optminrecall", 0.01)
	aReq.OptMaxPrec = parms.Fval("optmaxprec", 0.95)
	aReq.OkToRun = false
	aReq.DoPreAnalyze = parms.Bval("dopreanalyze", false)
	aReq.AnalClassId = int16(parms.Ival("analclassid", AnalNoClassSpecified))
	aReq.AnalSplitType = int16(parms.Ival("analsplittype", 1))
	aReq.AnalTestPort = parms.Fval("analtestport", 0.15)

	if aReq.TrainInFi == "" && aReq.ModelFi == "" {
		checkClassifyFilesParms("Either training file or model file must be specified", true)
		return aReq
	}

	if aReq.TrainInFi != "" {
		if _, err := os.Stat(aReq.TrainInFi); os.IsNotExist(err) {
			fmt.Printf("ERROR: train file does not exist %s\n", aReq.TrainInFi)
			printClassifyFilesHelp()
			return aReq
		}
	}

	if aReq.TestInFi != "" {
		if _, err := os.Stat(aReq.TestInFi); os.IsNotExist(err) {
			fmt.Printf("ERROR: test file does not exist %s\n", aReq.TestInFi)
			printClassifyFilesHelp()
			return aReq

		}
	}

	if aReq.ClassInFi != "" {
		if _, err := os.Stat(aReq.ClassInFi); os.IsNotExist(err) {
			fmt.Printf("ERROR: Class file does not exist %s\n", aReq.ClassInFi)
			printClassifyFilesHelp()
			return aReq

		}
	}

	if aReq.ModelFi != "" && aReq.LoadModel == true {
		if _, err := os.Stat(aReq.ModelFi); os.IsNotExist(err) {
			fmt.Printf("ERROR: model file does not exist %s\n", aReq.ModelFi)
			printClassifyFilesHelp()
			return aReq
		}
	}
	aReq.OkToRun = true
	return aReq

}
