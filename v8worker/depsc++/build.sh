#!/bin/bash

g++ -c -O2 -Wall -std=c++11 v8binding.cpp -I/usr/local/include/v8 -I/usr/local/Cellar/v8/8.1.307.32/libexec/include
rm -f libdepsc++.a
ar -r libdepsc++.a v8binding.o
