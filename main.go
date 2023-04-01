/*
__      _________      __
\ \    / /_   _\ \    / /\
 \ \  / /  | |  \ \  / /  \
  \ \/ /   | |   \ \/ / /\ \
   \  /   _| |_   \  / ____ \
 _  \/   |_____|  _\/_/    \_\     _   _
| | | |          |  _ \           | | | |
| |_| |__   ___  | |_) | ___  __ _| |_| | ___  ___
| __| '_ \ / _ \ |  _ < / _ \/ _` | __| |/ _ \/ __|
| |_| | | |  __/ | |_) |  __/ (_| | |_| |  __/\__ \
 \__|_| |_|\___| |____/ \___|\__,_|\__|_|\___||___/
*/

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"

	"os"

	// "log"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	// "fyne.io/fyne/app"
	"launch/resources"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {

	App := app.New()
	mainWindow := App.NewWindow("LF Launcher")
	mainWindow.Resize(fyne.NewSize(700, 270))

	img := canvas.NewImageFromResource(fyne.NewStaticResource("bg.jpg", resources.ResourceBgJpg.StaticContent))

	progress := widget.NewProgressBar()
	textLabel := widget.NewLabel("LastFrontier.ru...")
	stack := fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		progress,
		textLabel,
	)
	stack.Resize(fyne.NewSize(300, 50))
	textLabel.Alignment = fyne.TextAlignCenter
	progress.SetValue(0.5)

	startBtns := container.NewGridWithColumns(2, widget.NewButton("OpenGL ver.", func() { startGameOpenGl() }), widget.NewButton("Update", func() { startUpdate() }))
	manageZone := container.NewGridWithRows(2, startBtns, stack)
	bottn := container.NewGridWithRows(2, widget.NewButton("START GAME", func() { startGameDirectX() }), manageZone)
	mainWindow.SetContent(container.NewGridWithRows(2, img, bottn))

	mainWindow.Show()
	App.Settings().SetTheme(theme.DarkTheme())
	App.Run()

}

func startUpdate() {
	fmt.Println("Start Update")
	/* downloadNewClient()
	toTemp()
	replaceTemp()
	startGameDirectX() */
}

func startGameDirectX() {
	// получаем путь к текущей директории
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Ошибка при получении текущей директории:", err)
		return
	}

	// собираем относительный путь к исполняемому файлу
	exePath := filepath.Join(currentDir, "LastFrontier.exe")

	// запускаем исполняемый файл
	cmd := exec.Command(exePath)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Ошибка при запуске приложения:", err)
		return
	}
}

func startGameOpenGl() {
	// получаем путь к текущей директории
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Ошибка при получении текущей директории:", err)
		return
	}

	// собираем относительный путь к исполняемому файлу
	exePath := filepath.Join(currentDir, "LastFrontier_OpenGL.exe")

	// запускаем исполняемый файл
	cmd := exec.Command(exePath)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Ошибка при запуске приложения:", err)
		return
	}
}

func replaceTemp() {
	// Получаем текущую директорию
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Путь к папке temp
	tempDir := filepath.Join(dir, "temp")

	// Получаем список файлов и папок в папке temp
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		panic(err)
	}

	// Перебираем все файлы и папки в папке temp
	for _, file := range files {
		// Путь к файлу или папке в папке temp
		src := filepath.Join(tempDir, file.Name())

		// Путь к файлу или папке в текущей директории
		dst := filepath.Join(dir, file.Name())

		// Если файл или папка с таким именем уже есть в текущей директории, удаляем его
		if _, err := os.Stat(dst); err == nil {
			err = os.RemoveAll(dst)
			if err != nil {
				panic(err)
			}
		}

		// Перемещаем файл или папку в текущую директорию
		err = os.Rename(src, dst)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Все файлы из папки temp успешно перемещены в текущую директорию.")

	// Удаляем папку temp
	err = os.RemoveAll(tempDir)
	if err != nil {
		panic(err)
	}
}

func toTemp() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	tempDir := filepath.Join(dir, "temp")
	if err := os.Mkdir(tempDir, 0755); err != nil {
		panic(err)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".zip" {
			if err := unzip(path, tempDir); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	err = os.Remove("1233.zip")
	if err != nil {
		panic(err)
	}
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
			continue
		}

		if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadNewClient() {
	fileURL := "https://drive.google.com/u/0/uc?id=1AM9xNOXtM5ge0gnHPC5-UjpP-9TJfKil&export=download&confirm=no_antivirus"

	fileName := "LF-Client.zip"

	response, err := http.Get(fileURL)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Файл %s успешно загружен\n", fileName)
}
