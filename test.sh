#!/bin/bash
echo -n 'Go: '
./getver golang.org/project || echo
echo -n 'Python 2: '
./getver -u 2 python.org || echo
echo -n 'Python 3: '
./getver -u 1 python.org || echo
echo -n 'Rust: '
./getver rust-lang.org || echo
echo -n 'Grails: '
./getver -d 2 -u 1 --sort grails.org || echo
echo -n 'Groovy: '
./getver -d 2 -u 2 groovy-lang.org || echo
