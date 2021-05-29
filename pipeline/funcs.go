package pipeline

import (
	"fmt"
	"io"
	"os"
)

/************************************************
	Windows special symbols
	Use any character, except for the following:
	The following reserved characters:
		< (less than)
		> (greater than)
		: (colon)
		" (double quote)
		/ (forward slash)
		\ (backslash)
		| (vertical bar or pipe)
		? (question mark)
		* (asterisk)
		\0 Integer value zero, sometimes referred to as the ASCII NUL character.

		\1-\31 Characters whose integer representations are in the range from 1 through 31
	 https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file#file-and-directory-names

Characters from 1 through 31, except TAB and NEW LINE	remove
TAB (\t)	replace to " " (space)
NEW LINE (\n)	replace to " " (space)
/	replace to - (dash)
\	replace to - (dash)
:	replace to _ (underscore)
*	replace to _ (underscore)
?	remove
<	replace to _
>	replace to _
"	replace to ' (single quote)
 ************************************************/
func safeString(str string) string {
	res := ""

	for _, c := range str {
		switch c {
		case ' ', '\t', '\n':
			res += "_"
			continue

		case '\\', '/':
			res += "-"
			continue

		case ':', '*':
			res += "_"
			continue

		case '<':
			res += "["
			continue

		case '>':
			res += "]"
			continue

		case '?':
			continue

		case '"':
			res += "'"
			continue
		}

		if c <= 31 {
			continue
		}

		res += string(c)
	}

	if res == "." {
		return "_"
	}

	if res == ".." {
		return "__"
	}

	return res
}

func createDirWithPerm(dir string, perm os.FileMode) error {
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("Failed create directory.\n - directory: %s\n - error: %s", dir, err)
	}

	return nil
}

func createDir(dir string) error {
	return createDirWithPerm(dir, 0777)
}

// func mustRmDir(dir string) {
// 	if err := os.RemoveAll(dir); err != nil {
// 		log.Fatalf("can't delete directory %s: %s", dir, err)
// 	}
// }

func copyFileWithPerm(src, dest string, perm os.FileMode) error {
	copy := func() error {
		var err error
		var srcFile *os.File
		var dstFile *os.File

		if srcFile, err = os.Open(src); err != nil {
			return err
		}
		defer srcFile.Close()

		if dstFile, err = os.Create(dest); err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err = io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		return nil
	}

	if err := copy(); err != nil {
		return fmt.Errorf("Failed to copy file.\n - source: %s\n - destination: %s\n - error: %s", src, dest, err)
	}

	if err := os.Chmod(dest, perm); err != nil {
		return fmt.Errorf("Failed to chnage permission\n - file: %s\n - error: %s", dest, err)
	}

	return nil
}

func deletDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf(`can't remove directory "%s":\n  %s`, dir, err)
	}
	return nil
}

func isFileExists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}
