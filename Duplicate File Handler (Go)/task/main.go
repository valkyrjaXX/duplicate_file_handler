package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Directory is not specified")
		return
	}

	root := os.Args[1]

	fileFormat := readFileFormat()
	filesByPath := searchFiles(root, fileFormat)

	sortOption := readSortOption()

	filesSize, filesBySize := groupFilesBySize(filesByPath)
	sortFilesBySize(filesSize, filesBySize, sortOption)

	readForDuplicatesOption := readForDuplicatesOption()
	filesBySizeAndHash := groupFilesBySizeAndHash(filesBySize, readForDuplicatesOption)
	sortedFiles := sortFilesByHash(filesSize, filesBySizeAndHash, sortOption)

	deleteFilesOption := readForDeleteOption()
	filesToDelete := readForNumbersToDelete(deleteFilesOption, len(sortedFiles))
	deletedFiles := deleteFiles(sortedFiles, filesToDelete)
	calculateTotalDeletedSum(filesByPath, deletedFiles)

}

func readFileFormat() string {
	fmt.Println("Enter file format:")
	var fileFormat string
	fmt.Scanf("%s", &fileFormat)
	return fileFormat
}

func searchFiles(root, fileFormat string) map[string]int64 {
	filesByPath := make(map[string]int64)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if fileFormat == "" {
			filesByPath[path] = info.Size()
			return nil
		}

		if strings.HasSuffix(path, fileFormat) {
			filesByPath[path] = info.Size()
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error while walking through the directory:", err)
		return nil
	}

	return filesByPath
}

func groupFilesBySize(filesBySize map[string]int64) ([]int64, map[int64][]string) {
	var keys []int64

	result := make(map[int64][]string)
	for path, fileSize := range filesBySize {
		if _, ok := result[fileSize]; ok {
			result[fileSize] = append(result[fileSize], path)
		} else {
			keys = append(keys, fileSize)
			result[fileSize] = []string{path}
		}
	}

	return keys, result
}

func readSortOption() int {
	fmt.Println("Size sorting options:")
	fmt.Println("1. Descending")
	fmt.Println("2. Ascending")

	var sortOption int

	for {
		fmt.Scanf("%d", &sortOption)
		if sortOption == 1 || sortOption == 2 {
			break
		}
		fmt.Println("Wrong option")
	}

	return sortOption
}

func sortFilesBySize(filesSize []int64, filesBySize map[int64][]string, sortOption int) {
	if sortOption == 1 {
		sort.Slice(filesSize, func(i, j int) bool {
			return filesSize[i] >= filesSize[j]
		})
	} else {
		sort.Slice(filesSize, func(i, j int) bool {
			return filesSize[i] <= filesSize[j]
		})
	}

	for _, k := range filesSize {
		fmt.Println(k, " bytes")

		for _, path := range filesBySize[k] {
			fmt.Println(path)
		}
	}
}

func readForDuplicatesOption() bool {
	fmt.Println("Check for duplicates?")

	var checkForDuplicates string

	for {
		fmt.Scanf("%s", &checkForDuplicates)
		if checkForDuplicates == "yes" || checkForDuplicates == "no" {
			break
		}

		fmt.Println("Wrong option")
	}

	return checkForDuplicates == "yes"
}

func groupFilesBySizeAndHash(filesBySize map[int64][]string, readForDuplicatesOption bool) map[int64]map[string][]string {
	if !readForDuplicatesOption {
		return nil
	}

	var keys []int64
	filesBySizeAndHash := make(map[int64]map[string][]string)
	for size, files := range filesBySize {
		filesByHash := make(map[string][]string)
		for _, path := range files {
			hash, err := readFileHash(path)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			if _, ok := filesByHash[hash]; ok {
				filesByHash[hash] = append(filesByHash[hash], path)
			} else {
				filesByHash[hash] = []string{path}
			}
		}

		for hash, paths := range filesByHash {
			if len(paths) == 1 {
				delete(filesByHash, hash)
			}
		}

		if len(filesByHash) > 0 {
			keys = append(keys, size)
			filesBySizeAndHash[size] = filesByHash
		}
	}

	return filesBySizeAndHash
}

func readFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err = io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func sortFilesByHash(filesSize []int64, filesBySizeAndHash map[int64]map[string][]string, sortOption int) []string {
	if sortOption == 1 {
		sort.Slice(filesSize, func(i, j int) bool {
			return filesSize[i] >= filesSize[j]
		})
	} else {
		sort.Slice(filesSize, func(i, j int) bool {
			return filesSize[i] <= filesSize[j]
		})
	}

	var count int
	var sortedFiles []string
	for _, k := range filesSize {
		fmt.Println(k, " bytes")

		for hash, paths := range filesBySizeAndHash[k] {
			fmt.Println("Hash: ", hash)

			for _, path := range paths {
				count++
				sortedFiles = append(sortedFiles, path)
				fmt.Println(fmt.Sprintf("%d. %s", count, path))
			}
		}
	}

	return sortedFiles
}

func readForDeleteOption() bool {
	fmt.Println("Delete files?")

	var deleteFiles string

	for {
		fmt.Scanf("%s", &deleteFiles)
		if deleteFiles == "yes" || deleteFiles == "no" {
			break
		}

		fmt.Println("Wrong option")
	}

	return deleteFiles == "yes"
}

func readForNumbersToDelete(deleteFilesOption bool, numOfFiles int) []int {
	if !deleteFilesOption {
		return nil
	}

	for {
		in := bufio.NewReader(os.Stdin)
		filesToDelete, err := in.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		fileIndexesToDelete := strings.Split(filesToDelete[0:len(filesToDelete)-1], " ")

		var values []int
		for _, fileIndex := range fileIndexesToDelete {
			idx, err := strconv.Atoi(fileIndex)
			if err != nil {
				fmt.Println("Wrong format")
				break
			}

			if idx > numOfFiles {
				fmt.Println("Wrong format")
				break
			}

			values = append(values, idx)
		}

		if len(values) != 0 {
			return values
		}
	}
}

func deleteFiles(files []string, filesToDelete []int) []string {
	var deletedFiles []string
	for _, fileIdx := range filesToDelete {
		path := files[fileIdx-1]

		err := os.Remove(path)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		deletedFiles = append(deletedFiles, path)
	}

	return deletedFiles
}

func calculateTotalDeletedSum(filesByPath map[string]int64, deletedFiles []string) {
	var sum int64
	for _, file := range deletedFiles {
		if size, ok := filesByPath[file]; ok {
			sum += size
		}
	}

	fmt.Println(fmt.Sprintf("Total freed up space: %d bytes", sum))
}
