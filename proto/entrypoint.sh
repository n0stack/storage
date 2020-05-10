#!/bin/bash
set -x
set -e

# echo 'set --doc_out if you want api docs'

for pd in `find /src -maxdepth 1 -mindepth 1 -type d`
do
  package=`basename $pd`
  echo "{}" > $pd/$package.swagger.json

  for d in `find $pd -name '*.proto' | xargs dirname | uniq`
  do
    name=${d#\/src\/}
    protoc \
      -I/src \
      -I/opt/protoc/include \
      --go_out=plugins=grpc:/go/src \
      --grpc-gateway_out=/go/src \
      --swagger_out=logtostderr=true,allow_merge=true,merge_file_name=$package:$pd \
      $* \
      $d/*.proto
      # --doc_opt=markdown,${name//\//_}.md \
  done
done
