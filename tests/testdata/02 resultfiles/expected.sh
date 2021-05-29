#!/bin/sh

mkdir -p MacOS
echo "file 1" > MacOS/Binary
chmod 755 MacOS/Binary

mkdir -p 02.1_SrcArray
echo "02 src array 1" > 02.1_SrcArray/02_src_array1.txt
echo "02 src array 2" > 02.1_SrcArray/02_src_array2.txt
chmod 664 02.1_SrcArray/*

mkdir -p 02.2_SrcArray
echo "02 src array 1" > 02.2_SrcArray/02_src_array1.txt
echo "02 src array 2" > 02.2_SrcArray/02_src_array2.txt
chmod 664 02.2_SrcArray/*

mkdir -p 03_glob
echo "glob 1" > 03_glob/glob-file1.txt
echo "glob 2" > 03_glob/glob-file2.txt
echo "glob 1" > 03_glob/glob-file1.md
chmod 664 03_glob/*

mkdir -p "04_subdirs"
echo "subdirs 1" > 04_subdirs/file_1.txt
echo "subdirs 2" > 04_subdirs/file_2.txt
chmod 664 04_subdirs/*

mkdir -p "05_subdirs"
echo "subdirs 1" > 05_subdirs/file_1.txt
echo "subdirs 2" > 05_subdirs/file_2.txt
chmod 664 05_subdirs/*


echo "05_rename" > 05_rename.txt
chmod 664 05_rename.txt

mkdir -p 06_permissions
echo "06_permissions_753" > 06_permissions/06.1_permissions_753
chmod 753 06_permissions/06.1_permissions_753
echo "06_permissions_753" > 06_permissions/06.2_permissions_753
chmod 753 06_permissions/06.2_permissions_753

echo "07_env_BIN_DIR" > MacOS/07_env_BIN_DIR.txt
chmod 664 MacOS/07_env_BIN_DIR.txt
