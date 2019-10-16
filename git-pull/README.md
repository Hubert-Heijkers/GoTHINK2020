This code is part of the TM1 SDK Hands-On Lab

The 'git-pull' app is used to demo how one could deploy a version of a TM1 model, in this case one that was previously build using the 'builder', from source using the github.com/hubert-heijkers/tm1-model-northwind GIT repository. The git-pull app is itself uses the REST API from, in this particular case, Go.

The app initializes GIT, tying the NorthWind database to the tm1-model-northwind repository, and subsequently pulls the code and applies it to the, presumably empty, server. After it is done all the artifacts from the source are updated accordingly and any pre and post pull steps have been executed.
