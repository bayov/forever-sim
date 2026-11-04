"""mkmany.py ROT TALENTS PREFIX 'name knob=v ...' ...: several requests at .ps/PREFIXname.json"""
import subprocess, sys
rot, talents, prefix = sys.argv[1:4]
for spec in sys.argv[4:]:
    parts = spec.split()
    subprocess.run([sys.executable, '.ps/mk.py', f'.ps/{prefix}{parts[0]}.json', rot, talents] + parts[1:], check=True)
    print(f'.ps/{prefix}{parts[0]}.json')
