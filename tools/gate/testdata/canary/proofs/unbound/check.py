import pathlib
import sys

if 'bad' in pathlib.Path(sys.argv[1]).read_text():
    print('A.rule: violation')
    sys.exit(1)
