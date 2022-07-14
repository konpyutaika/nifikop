#! /bin/sh

# run helm dependency update and expand all tarballs
(rm -rf charts && helm dep update && cd charts && for filename in *.tgz; do tar -xf "$filename" && rm -f "$filename"; done;)
