echo "Start building..."
python3 build.py -p cli/ -o build/ -n "tycl-{os}-{arch}"
echo "Compleate."
echo "Binary files available in 'build/'"