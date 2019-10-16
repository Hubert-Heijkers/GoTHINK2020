This code is part of the TM1 SDK Hands-On Lab

The 'watcher' is used to show how one can trail the transaction log using OData's support for delta's. Trailing the transaction log, or the message log for that matter, enables developers to programmatically react to changes and/or events happening in the system. In this case we'll simply dump any changes, in the context of this lab made by the loader while loading data into our Sales cube, to the console. The watcher code itself uses the REST API from, in this particular case, Go.
