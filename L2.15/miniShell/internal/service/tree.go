package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DirectoryTree struct {
	root   node
	curDir node
}

type node struct {
	absPath string
	isDir    bool
	parant   *node
	children []*node
}

func newNode(absPath string, isDir bool, parant *node) node{
	return node{absPath: absPath, isDir: isDir, parant: parant}
}

func (n *node) setCildren(ch []*node) {
	n.children = ch
}

// if there are no parametrs -- like NewDirectoryTree()
// this function will return DirectoryTree with curDir = (directory, where were this function called)
// for exp if your binFile located byPath /MyProj/bins, NewDirectoryTree() will return tree with root==curDir==bins
// so it`s better to wtite the global Path or relative path, from /MyProj/bins to directory, where you want to create tree
func NewDirectoryTree(args ...string) (*DirectoryTree, error) {
	curDirPath, err := os.Getwd()

	if err != nil {
		return nil, 
		fmt.Errorf("Failed to load the absolute path to the directory containing the executable go file, err = %s", err.Error())
	}

	if len(args) == 0 {
		cn := newNode(curDirPath, true, nil)
		return &DirectoryTree{curDir: cn, root:  cn}, nil
	}

	if len(args) > 1 {
		return nil,
		fmt.Errorf("function NewDirectoryTree can take no more than 1 argument")
	}

	path := args[0]
	fi, err := os.Stat(path) // вот тут проверь
	if err != nil || !fi.IsDir(){
		return nil,
		fmt.Errorf("can`t open directory by path %s, err = %w", path, err)
	}
	var absPath string
	if filepath.IsAbs(path) {
		absPath = path
	} else {
		locAbsPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		absPath = locAbsPath
	}
	cn := newNode(absPath, true, nil)
	
	return &DirectoryTree{root: cn, curDir: cn}, nil
}


// === utils ===
func pathToNodeName(path string) string {
	arr := strings.Split(path, string(os.PathSeparator))
	return arr[len(arr)-1]
}

func giveChildren(path string, parant *node) ([]*node, error) {
	dirInfo, err := os.Stat(path)
	if err != nil {
		return nil,
		fmt.Errorf("can`t open file/dir by path %s", path)
	}

	if !dirInfo.IsDir() {
		return nil,
		fmt.Errorf("file by path %s shold be directory? but it isn`t a directory", path)
	}

	entries, err := os.ReadDir(path)
    if err != nil {
        return nil, fmt.Errorf("unexpected error = %w", err) // we do not expect an error here, because we checked all errors higher
    }

	childrens := make([]*node, 0, len(entries))
    for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		n := newNode(fullPath, entry.IsDir(), parant)
		childrens = append(childrens, &n)
    }

    return childrens, nil
}


//DEV
/*
os.Stat(path) в giveChildren избыточен: Ты уже проверял, что путь — директория в NewDirectoryTree (или предполагаешь), 
но если вызывать giveChildren отдельно — ок. Однако os.ReadDir сам проверит, и если ошибка — вернёт её. Дубли добавляют overhead.
Фикс: Убрать Stat в giveChildren, полагаться на ReadDir. Если нужно info — используй entry.Type() из entries (хотя для IsDir ок).
*/