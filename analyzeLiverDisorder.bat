::rm classifyFiles.exe
::go build src/classifyFiles.go
classifyFiles -train=data/liver-disorder.test.csv -test=data/liver-disorder.train.csv -maxBuck=500 -testout=tmpout/liver-disorder.test.csv -detToStdOut=false  -doPreAnalyze=true -AnalSplitType=1 -AnalClassId=1  -AnalTestPort=0.1

