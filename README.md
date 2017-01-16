# ML Quantized classifier in GO#

A general purpose, high performance machine learning
classifier.  Tests are:

   *  ASL Sign language Gesture recognition 
   *  Tests classify diabetes 
   *  predict death or survival of Titanic passengers. 

Also includes TensorFlow implementation of classifiers 
using the same data to compare the run-time performance
and classification accuracy.  

My design for Quantized classifiers was inspired by 
design elements in KNN, Bayesian and SVM engines. 
A key design goal was a faster mechanism to identify 
similarity for a given feature. 

  In KNN we find similar records for a given feature by finding 
  those with the most similar value which works but consumes
  a lot of space and run-time.   In Quantized approach 
  we look at the range of data and attempt to 
  group data based on similar values.  EG: if a given
  feature has a value from 0.0 to 1.0 then a 10 bucket system 
  could consider all records that have a value from 0 to 0.1 as 
  similar. Those from 0.1 to 0.2 are similar, etc.  Rather than 
  keeping all the training records we only need to keep 
  the statistics for have many of the records in a given
  bucket belong to 
  each class which we can then use to compute a base probability
  by feature by bucket by class. Applying this across active 
  features gives us a set of probabilities that can be combined
  using ensemble techniques into the probability a given
  row would belong to any of the classes.  
  Quantizing the data allows a small memory foot print 
  with fast training without the KNN
  need of keeping all the training records in memory. Retaining
  only the statistics allows 
  very large training sets. The trade off is that it looses some of 
  KNN ability to adjust the number of closest neighbors considered.
  but training is so fast that the quanta size can be adjusted quickly
  and memory use is so much smaller that we can afford to keep
  multiple models with different quanta sizes loaded and updated
  simultaneously. 


###ASP (American Sign Language) Gesture classifier###
This engine started as a classifier designed to classify Static Gestures for VR with the idea we may be able to produce a useful tool for classifying  ASL using VR input devices.  That is still a primary focus but the core algorithms can be more broadly applied.

See **Overview.pdf** in this repository for conceptual overview of
the approach when using this kind of classifier for gesture recognition.

This repository includes code written to test ideas for static gesture recognition. 

It also includes samples of the classifiers in python that cope
well with smaller training data sets and demonstrate using 
the Quantized classifier approach.  They also handle
massive training data sets with minimal memory.    



 * Version: 0.1
 * License: (MIT) We do sell consulting services http://BayesAnalytic.com/contact
 * Dependencies: 
    - GO Code was built using version 1.7.3 windows/amd 64
    - Python code: Was tested with Python 3.5.2 64 bit
    - TensorFlow: Lots of crazy dependencies See: tlearn/tensflowReadme.docx 

### How to Use ###
  * **python quant_filt.py** - Runs test on gesture classification data.
    Shows how quantized concept can be used to implement a
    splay like search tress.  The more quant buckets used 
    the more precise.  This is an alternative to the probability
    model and can provide superior results in some instance.
  
  * **python quant_prob.py** - Runs a test on gesture classification data
    demonstrates quantized probability theory in smallest possible 
    piece of python code.  A more complete version is implemented 
    in classify.go 
    
  * **makeGO.bat** - if you have GO set up then open a command line at
    the base directory containing makeGO.bat and run it.   It should
    setup GOHOME and build the executable files. Tested on windows 10.
    
  * **splitData.bat** - Creates sub .train.csv and test.csv files for the files
    used in the GO classifier tests. Uses splitCSVFile.exe which is built
    by makeGo. 
    
  * **setGoEvn.bat** - will set the GOHOME directory to current working directory
    in a command prompt.
    
  * **go build src/classifyTest.go**
    builds executable classifyTest from GO source. 
    
  * **classifyTest data/diabetes...  data/daibeaat **
    will run the GO based classifier built in GO using
    the first named file for training and the second named
    file for testing will print out results of how well classification
    matches actual source data class.
    
  * **classifyTest data/titantic.train.csv data/titantic.test.csv**
    will run the GO based classifier against the two input files
    this test attempts to predict mortality and will print out
    quality of predictions from classifier compared to known
    result. 
    

### Basic Contents ###
#### GO Based Classifier ####
  src/classifyTest.go
  
  src/csvInfoTest.go
  
  src/splitCSVFile.go 
  
  src/qprob/classify.go
  
  src/qprob/csvInfo.go
  
  src/qprob/util.go
  
  
  
 
  
#### Idea Test Sample Code ####
* **quant_filt.py**  - Machine learning Quantized filter classifier.  This system can provide  
   fast classification with moderate memory use and is easy to see how likely the match is to
   be accurate.

* **quant_prob.py** - Machine learning Quantized probability classifier. Not quite as precise under
   some conditions and quant_filt.py but it can cope with greater amounts of training noise while
   still delivering good results with moderate amounts of training data.  
 

####DATA FILES####
 * **data/data-sources.txt** - Explains sources for the included data files
   some data files are not included and will have to be donwloaded from
   those sources if the usage license was unclear or restrictive.
   
 * **data/train/gest_train_ratio2.csv** - Input training data used for these tests.  We need thousands additional training samples feel free to volunteer after your read overview.pdf in this repository.


####TensorFlow###
 One of the goals this project is to test some
 capabilities of tlearn and TensorFlow using the 
 same data sets.   The assertion is that the 
 tensorflow approach should run faster and require
 less code while producing higher quality classification
 results than my quantized classifier. 
 
* **tlearn/tensFlowReadme.docx** - Notes I made while getting tensor flow running on my windows laptop.


* **tlearn/simple_gestures.py** - sample of reading CSV to  train TensorFlow Model.
   Unfortunately this program while it runs does a pour job of classification. I think
   this is the result of insufficient training data but there is a chance that I still have
   a bug in the interface to TensorFlow.




### How do I get set up? ###

* Summary of set up
* Configuration
* Dependencies - Tested on Python 3.5 on Windows 10.
* Database configuration
* How to run tests
* Deployment instructions

### Contribution guidelines ###

* Writing tests
* Code review
* Other guidelines

### Who do I talk to? ###

* Repo owner Joseph Ellsworth
* We do sell consulting services http://BayesAnalytic.com/contact


#TODO#

##Actions for Both quant_prob and quant_filt##
* Update QuantProb to properly scale buckets to cope with outliers
* Utility to split input files into separate sets.
* Update diabetes test to reserve some data for train, some for test
  and to call those two. 

  
  
* Modify Quant_prob run as server handler. 
  * Method will use main as data set name unless &dset is specified.
  * Each named data set is unique and will not conflict with others.
  * Method to add to training data set with POST BODY
  * Method to add to training data set with URI to fetch.
    * Allow the system to skip every N records to reserve for 
    * testing.
  * Method to classify values in file at URI 
    * Allow &testEvery=X to only test every Nth
      item.  This is to support testing.     
  * Method to classify with POST multiple lines.
  * Method to classify with GET for single set of features.
  * Allow number of buckets to be set by column name
  * allow column name to be set map direct to bucket id

  *    
* Produce a version for text parsing that computes position
    indexed position of all words where each unique word gets 
    a column number.   Then when building quantized tree 
    lookup of the indexed position for that word  treat the word 
    index as the bucketId or possibly as columnNumber need to think
    that one through buck as a bucket id seems to make most sense
    nd then 
    treat all the other features as empty. So the list of cols
    may grow to several million but will only contain the hashed
    classes for those value. Allow system to pass in a list
    of columns n the CSV to index as text.  This would not 
    effectively use word ordering but we could use quantized buckets
    for probability of any word being that text in text string so
    a word like "the" that may occur 50 times would occur in a different
    bucket when it is repeated many times. 
  * Only include detail probs if requested.
  * Choose column to use as class

  * Test with following data
     * Diabetes classification
     * 
     

* [Learn Markdown](https://bitbucket.org/tutorials/markdowndemo)
