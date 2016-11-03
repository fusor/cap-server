#!/bin/bash
registry="$1"
nulecule_name="$2"
nulecule_dir=$HOME/nulecules/$registry/$nulecule_name

echo "============================================================"
echo "CAP RUNNING NULECULE:"
echo "REGISTRY: $registry"
echo "NULECULE_NAME: $nulecule_name"
echo "NULECULE_DIR: $nulecule_dir"
echo "============================================================"

script -c " cd $nulecule_dir && \\
  atomic run $registry/$nulecule_name -v \\
  -a $nulecule_dir/answers.conf.gen --provider openshift" /dev/null
