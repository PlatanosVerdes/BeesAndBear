import argparse, os, random

parser = argparse.ArgumentParser()
parser.add_argument("-p", "--producer", help="Producer name", default="Baloo", type=str)
parser.add_argument("-c", "--consumer", help="Consumer name file", default="BEENAMES", type=str)
parser.add_argument("-n", "--number", help="Number of consumers", default=3, type=int)
parser.add_argument("-o","--output", help="Output file", default="start_program.bat", type=str)
parser.add_argument("-r", "--run", help="Run the generated .bat", default=False, type=bool)
args = parser.parse_args()

def main():

    out = rf'''@echo off

:: BEAR
start /B go run ''' + os.getcwd() + r'''\bear.go ''' + args.producer + r'''

:: BEES
'''
    for i in range(args.number-1):
        random_name = random.choice(list(open("BEE_NAMES", encoding="utf8"))).split("\n")[0]
        out += r'''start /B go run ''' + os.getcwd() + r'''\bees.go ''' + random_name + r'''
'''

    random_name = random.choice(list(open("BEE_NAMES", encoding="utf8"))).split("\n")[0]
    out += r'''go run ''' + os.getcwd() + r'''\bees.go ''' + random_name

    generator = open(os.getcwd() + f"\{args.output}", "w")
    generator.write(out)
    generator.close()

    if args.run:
        os.system('cmd /k ".\start_program.bat"')

if __name__ == "__main__":
    main()