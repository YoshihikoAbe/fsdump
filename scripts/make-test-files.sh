#!/bin/bash

for i in {1..100}; do
    echo -n $i "test" > testdata/$i
done