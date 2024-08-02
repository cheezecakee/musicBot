# musicBot

Supports wake word detection.
Current working words are:
- play
- pause
- resume
- skip 
- back
- queue

Works best with [vosk-model-en-us-0.42-gigaspeech](https://alphacephei.com/vosk/models)

Only been test on windows using powershell 7. 
Might need to change the exec file and nlp ".h" ".dll/.so" files according to [vosk](https://github.com/alphacep/vosk-api/releases) 

# Acknowledgments

The speech-text builds upon code and ideas from the following sources:

- [inevolin](https://github.com/inevolin/DiscordEarsGo) - Initial code for integrating Vosk with a Go application. This code has been modified significantly to fit the specific needs of this project.



