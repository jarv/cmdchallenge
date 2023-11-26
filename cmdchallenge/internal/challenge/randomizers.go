package challenge

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path"
	"strconv"
)

var ErrRandomizerNotExist = errors.New("randomizer does not exist")
var ErrRandomizerNotImplemented = errors.New("randomizer not implemented")

type Randomizer struct {
	log *slog.Logger
	ch  *Challenge
}

type RandomizerFuncType func(*Randomizer) ([]string, error)

var rndTable = map[string]RandomizerFuncType{
	"count_files":                          (*Randomizer).rndCountFiles,
	"count_string_in_line":                 (*Randomizer).rndCountStringInLine,
	"dirs_containing_files_with_extension": (*Randomizer).rndDirsContainingFilesWithExtension,
	"find_primes":                          (*Randomizer).rndFindPrimes,
	"find_tabs_in_a_file":                  (*Randomizer).rndFindTabsInAFile,
	"list_files":                           (*Randomizer).rndListFiles,
	"nested_dirs":                          (*Randomizer).rndNestedDirs,
	"sum_all_numbers":                      (*Randomizer).rndSumAllNumbers,
	"search_for_files_containing_string":   (*Randomizer).rndSearchForFilescontainingString,
	"oops_list_files":                      (*Randomizer).rndOopsListFiles,
}

func NewRandomizer(log *slog.Logger, ch *Challenge) *Randomizer {
	return &Randomizer{log, ch}
}

func (r *Randomizer) RunRandomizer() ([]string, error) {
	if err := os.Chdir(path.Join("/var/challenges", r.ch.Dir())); err != nil {
		return nil, err
	}

	rndFn, exists := rndTable[r.ch.Slug()]
	if !exists {
		return nil, ErrRandomizerNotExist
	}

	rndResult, err := rndFn(r)
	if err != nil {
		return nil, err
	}

	return rndResult, nil
}

const (
	randMin = 10
)

func rndNum(randMax int) int {
	if randMax <= randMin {
		panic("invalid randmax")
	}

	return rand.Intn(randMax-randMin) + randMin //#nosec G404
}

func touchFile(fname string) error {
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func appendLine(fname, line string) error {
	const mode = 0644

	f, err := os.OpenFile(fname, os.O_APPEND|os.O_WRONLY|os.O_CREATE, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(line + "\n"); err != nil {
		return err
	}
	return nil
}

func writeLine(fname, line string) error {
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(line + "\n"); err != nil {
		return err
	}
	return nil
}

func (r *Randomizer) intFromFirstLine() int {
	v, err := strconv.Atoi(r.ch.ExpectedLines()[0])
	if err != nil {
		panic(err)
	}
	return v
}

func (r *Randomizer) strFromFirstLine() string {
	v := r.ch.ExpectedLines()[0]
	return v
}

func (r *Randomizer) rndCountFiles() ([]string, error) {
	numNewFiles := rndNum(20)

	for i := 0; i < numNewFiles; i++ {
		if err := touchFile(fmt.Sprintf("rand-%d", i)); err != nil {
			return nil, err
		}
	}
	return []string{strconv.Itoa(r.intFromFirstLine() + numNewFiles)}, nil
}

func (r *Randomizer) rndCountStringInLine() ([]string, error) {
	numLines := rndNum(20)

	for i := 0; i < numLines; i++ {
		if err := appendLine("access.log", "GET"); err != nil {
			return nil, err
		}
	}

	return []string{strconv.Itoa(r.intFromFirstLine() + numLines)}, nil
}

func (r *Randomizer) rndDirsContainingFilesWithExtension() ([]string, error) {
	var newExpectedLines = make([]string, len(r.ch.ExpectedLines()))
	copy(newExpectedLines, r.ch.ExpectedLines())

	for i := 0; i < rndNum(30); i++ {
		dname := fmt.Sprintf("a/b/c/%d", i)
		fname := path.Join(dname, "some-file.tf")
		if err := os.MkdirAll(dname, os.ModePerm); err != nil {
			return nil, err
		}

		if err := touchFile(fname); err != nil {
			return nil, err
		}

		newExpectedLines = append(newExpectedLines, dname)
	}
	return newExpectedLines, nil
}

func (r *Randomizer) rndFindPrimes() ([]string, error) {
	primes := [...]int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59,
		61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107, 109, 113, 127, 131, 137,
		139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193, 197, 199}

	rnd := rndNum(len(primes))
	for i := 0; i < rnd; i++ {
		if err := appendLine("random-numbers.txt", strconv.Itoa(primes[i])); err != nil {
			return nil, err
		}
	}

	return []string{strconv.Itoa(r.intFromFirstLine() + rnd)}, nil
}

func (r *Randomizer) rndFindTabsInAFile() ([]string, error) {
	rnd := rndNum(20)

	for i := 0; i < rnd; i++ {
		if err := appendLine("file-with-tabs.txt", "\t"); err != nil {
			return nil, err
		}
	}

	return []string{strconv.Itoa(r.intFromFirstLine() + rnd)}, nil
}

func (r *Randomizer) rndListFiles() ([]string, error) {
	var newExpectedLines = make([]string, len(r.ch.ExpectedLines()))
	copy(newExpectedLines, r.ch.ExpectedLines())
	numNewFiles := rndNum(20)

	for i := 0; i < numNewFiles; i++ {
		fname := fmt.Sprintf("rand-%d", i)
		if err := touchFile(fname); err != nil {
			return nil, err
		}
		newExpectedLines = append(newExpectedLines, fname)
	}

	return newExpectedLines, nil
}

func (r *Randomizer) rndNestedDirs() ([]string, error) {
	rnd := strconv.Itoa(rndNum(1000))

	if err := writeLine(".../  /. .the flag.txt", rnd); err != nil {
		return nil, err
	}

	return []string{rnd}, nil
}

func (r *Randomizer) rndSumAllNumbers() ([]string, error) {
	sum := r.intFromFirstLine()
	rnd := rndNum(1000)

	if err := appendLine("sum-me.txt", strconv.Itoa(rnd)); err != nil {
		return nil, err
	}

	return []string{strconv.Itoa(sum + rnd)}, nil
}

func (r *Randomizer) rndSearchForFilescontainingString() ([]string, error) {
	var newExpectedLines = make([]string, len(r.ch.ExpectedLines()))
	copy(newExpectedLines, r.ch.ExpectedLines())

	numNewFiles := rndNum(20)

	for i := 0; i < numNewFiles; i++ {
		fname := fmt.Sprintf("rand-%d", i)
		if err := writeLine(fname, "500"); err != nil {
			return nil, err
		}
		newExpectedLines = append(newExpectedLines, fname)
	}

	return newExpectedLines, nil
}

func (r *Randomizer) rndOopsListFiles() ([]string, error) {
	newLine := r.strFromFirstLine()
	fname := fmt.Sprintf("zzz-%d", rndNum(1000))
	if err := touchFile(fname); err != nil {
		return nil, err
	}

	return []string{newLine + " " + fname}, nil
}
