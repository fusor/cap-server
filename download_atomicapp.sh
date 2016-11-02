#!/bin/bash
registry="$1"
nulecule_name="$2"
nulecule_dir=$HOME/nulecules/$registry/$nulecule_name

echo "============================================================"
echo "CAP DOWNLOADING NULECULE"
echo "REGISTRY: $registry"
echo "NULECULE_NAME: $nulecule_name"
echo "NULECULE_DIR: $nulecule_dir"
echo "============================================================"

# Clean whatever is existing since atomic will not overwrite answers
sudo rm -rf $nulecule_dir
mkdir -p $nulecule_dir

# Need to wrap atomic calls in "script" to fake a tty
# atomic runs docker with the -t flag, which will break if this script
# is called from the go server

# HACK: Really need to clean this up and figure out a better way to script
# Multiple script commands in sequence within this script will not run
# sequentially! To make sure these commands execute in sequence, needed to
# run them all within the script wrapper. Obviously this is very ugly, need
# to figure out how to clean this up and fix it the proper way.
script -c "atomic run $registry/$nulecule_name \\
  --mode fetch --destination $nulecule_dir && \\
  pushd $nulecule_dir && \\
  atomic run $registry/$nulecule_name --mode genanswers && \\
  popd && sudo chown -R vagrant:vagrant $nulecule_dir" /dev/null
