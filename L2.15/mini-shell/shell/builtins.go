package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// RunBuiltin проверяет и выполняет команду, если она встроена
func RunBuiltin(args []string) (bool, error) {
	if len(args) == 0 {
		return false, nil
	}

	switch args[0] {
	case "cd":
		return true, builtinCD(args)
	case "pwd":
		return true, builtinPWD()
	case "echo":
		return true, builtinEcho(args)
	case "kill":
		return true, builtinKill(args)
	case "ps":
		return true, builtinPS()
	case "ls":
		return true, builtinLS(args)
	case "exit":
		os.Exit(0)
		return true, nil
	}
	return false, nil
}

// --- Реализации команд ---

func builtinCD(args []string) error {
	dir := ""
	if len(args) < 2 {
		userDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = userDir
	} else {
		dir = args[1]
	}
	return os.Chdir(dir)
}

func builtinPWD() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(dir)
	return nil
}

func builtinEcho(args []string) error {
	fmt.Println(strings.Join(args[1:], " "))
	return nil
}

func builtinKill(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: kill <pid>")
	}
	pid, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid pid: %s", args[1])
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// На Windows сигналы ограничены, поэтому используем Kill
	// Это универсальный способ убить процесс в Go
	return proc.Kill()
}

// builtinPS адаптирован под разные ОС
func builtinPS() error {
	if runtime.GOOS == "windows" {
		// На Windows самый простой способ получить список без тяжелых библиотек - вызвать tasklist
		// Мы эмулируем поведение, но под капотом используем инструменты ОС
		cmd := exec.Command("tasklist")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Для Linux/MacOS читаем /proc
	fmt.Println("PID\tCMD")
	files, err := os.ReadDir("/proc")
	if err != nil {
		// Fallback для MacOS, где /proc не смонтирован
		cmd := exec.Command("ps", "-e")
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}
		
		// Попытка прочитать имя процесса
		commPath := filepath.Join("/proc", f.Name(), "comm")
		content, err := os.ReadFile(commPath)
		if err != nil {
			continue
		}
		fmt.Printf("%d\t%s", pid, string(content))
	}
	return nil
}

// builtinLS - эмуляция команды ls на Go
func builtinLS(args []string) error {
	targetDir := "."
	showDetails := false

	// Простейший парсинг аргументов
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "l") {
				showDetails = true
			}
		} else {
			targetDir = arg
		}
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	if showDetails {
		// Формат списка (упрощенный аналог ls -l)
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			modTime := info.ModTime().Format(time.DateTime)
			size := info.Size()
			
			// Определение типа: D - dir, - - file
			typeChar := "-"
			if entry.IsDir() {
				typeChar = "d"
			}
			
			// Вывод: Type Permissions Size Date Name
			fmt.Printf("%s%s %8d %s %s\n", 
				typeChar, info.Mode().Perm(), size, modTime, entry.Name())
		}
	} else {
		// Простой вывод в строку
		var names []string
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			names = append(names, name)
		}
		fmt.Println(strings.Join(names, "  "))
	}

	return nil
}