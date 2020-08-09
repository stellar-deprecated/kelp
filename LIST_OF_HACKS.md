# LIST OF HACKS

## Awating v2.0

Incomplete list of hacks in the codebase that should be fixed before upgrading to v2.0 which will change the API to Kelp in some way

- LOH-1 - support backward-compatible case of not having any pre-specified function
- LOH-2 - support backward-compatible case of defaulting to "mid" price when left unspecified
- LOH-3 - we want to guarantee that the bot crashes if the errors exceed deleteCyclesThreshold, so we start a new thread with a sleep timer to crash the bot as a safety

## Workarounds

Incomplete list of workaround hacks in the codebase that should be fixed at some point