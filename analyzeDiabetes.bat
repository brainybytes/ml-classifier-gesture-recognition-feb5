:: Example of using the pre-analyze feature
:: to discover important features and important
:: data clusters within the important features.

classifyFiles -train=data/diabetes.train.csv -test=data/diabetes.test.csv -maxBuck=700 -testOut=tmpout/diabetes.test.csv  -detToStdOut=false -doPreAnalyze=true -AnalSplitType=1 AnalTestPort=0.1  -AnalClassId=1
