# concurrent-file-processing
 Concurrent text file read and write data(append), by N number of request concurrently;

## Problem statement
  App to read a text file (e.g, hello.txt) and write data in to the file(append), by N number of request concurrently. 
  
* Have to use message queue to avoid data loss and to improve the performance.
* Return the success response to each and every requests once the read or write operation done successfully. 

## Language

> GoLang - v1.11
