#!/bin/bash
echo -n 'Go: '
./getver golang.org
echo -n 'Python 2: '
./getver -u 2 python.org
echo -n 'Python 3: '
./getver -u 1 python.org
echo -n 'Rust: '
./getver rust-lang.org
echo -n 'Grails: '
./getver -d 2 -u 1 --sort grails.org
echo -n 'Groovy: '
./getver -d 2 -u 2 groovy-lang.org
