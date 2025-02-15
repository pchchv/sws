# SWS [![GoDoc](https://godoc.org/github.com/pchchv/sws?status.svg)](https://godoc.org/github.com/pchchv/sws) [![Go Report Card](https://goreportcard.com/badge/github.com/pchchv/sws)](https://goreportcard.com/report/github.com/pchchv/sws)

Static web server with live reloading.

## Install

```sh
go install github.com/pchchv/sws
```

## Usage

#### For hosting the current work directory:

```sh
sws s|serve
```

or 

```sh
sws s|serve <relative directory>
```

## Architecture
* First the content of the website is copied to a temporary directory, this is the _mirrored content_.
* Each mirror file is inspectd for type, if it is text/html, the `delta-streamer.js` script is injected.
* A web server is started, which hosts the _mirrored_ content.
* In turn, `delta-streamer.js` in turn sets up a websocket connection to the sws webserver.
* Еhe original file system is monitored, with any file changes:
  + the new file is copied to the mirror (including injections)
  + the file name is passed to the browser via websocket
* The `delta-streamer.js` script then checks if the current window origin is the updated file. If so, it reloads the page.
```
       ┌───────────────┐                                                 
       │ Web Developer │                                                 
       └───────┬───────┘                                                 
               │                                                         
       [writes <content>]                                                
               │                                                         
               ▼                                                         
 ┌─────────────────────────────┐        ┌─────────────────────┐          
 │ website-directory/<content> │        │ file system notify  │          
 └─────────────┬───────────────┘        └─────────┬───────────┘          
               │                                  │                      
               │                      [update mirrored content]          
               ▼                                  │                      
     ┌────────────────────┐                       │                      
     │ ws-script injector │◄──────────────────────┘                      
     └─────────┬──────────┘                                              
               │                                                         
               │                                                         
               ▼                                                         
   ┌────────────────────────┐                                            
   │ tmp-abcd1234/<content> │                                            
   └───────────┬────────────┘                                            
               │                                                         
       [serves <content>]                                                
               │                               ┌────────────────────────┐
               ▼                               │         Browser        │
┌──────────────────────────────┐               │                        │
│          Web Server          │               │  ┌────┐  ┌───────────┐ │
│ [localhost:<port>/<content>] │◄───[reload────┼─►│ ws │  │ <content> │ │
└──────────────────────────────┘     page]     │  └────┘  └───────────┘ │
                                               │                        │
                                               └────────────────────────┘
```
