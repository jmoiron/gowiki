#!/bin/bash

echo "package main" > data.go
echo "" >> data.go

echo "var t_files = map[string]string{" >> data.go
for f in templates/*; do
    name=${f#templates/}
    echo "  \"$name\": \`" >> data.go
    cat $f >> data.go
    echo "\`," >> data.go
    echo "" >> data.go
done
echo "}" >> data.go

echo "" >> data.go

echo "var s_files = map[string]string{" >> data.go
for f in static/*; do 
    echo "  \"$f\": \`" >> data.go
    cat $f |base64 >> data.go
    echo "\`," >> data.go
    echo "" >> data.go
done
echo "}" >> data.go
echo "" >> data.go

for f in static/*.md; do
    name=${f#static/}
    name=${name%.*}
    echo "var m_$name = \`" >> data.go
    cat $f >> data.go
    echo "\`" >> data.go
    echo "" >> data.go
done

