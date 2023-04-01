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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"launch/data"
	"launch/resources"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {

	App := app.New()
	mainWindow := App.NewWindow(data.WINDOW_NAME)
	mainWindow.Resize(fyne.NewSize(700, 270))

	// Изображение преобразовано в байты, чтобы не использовать лишний ресурс
	img := canvas.NewImageFromResource(fyne.NewStaticResource(data.IMG_NAME, resources.ResourceBgJpg.StaticContent))

	// Прогрессбар с инфо-текстом
	progressBar := widget.NewProgressBar()
	progressText := widget.NewLabel(data.PROGRESSBAR_TITLE_IDL)
	stack := fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		progressBar,
		progressText,
	)
	stack.Resize(fyne.NewSize(300, 50))
	progressText.Alignment = fyne.TextAlignCenter

	// Кнопки запуска Клиента на OpenGL и Запуска Обновления
	startBtns := container.NewGridWithColumns(2, widget.NewButton(data.BTN_START_OGL_TITLE,
		func() { startGameOpenGl(&mainWindow) }),
		widget.NewButton(data.BTN_UPDATE_TITLE,
			func() { startUpdate(progressBar, progressText, &mainWindow) }))
	manageZone := container.NewGridWithRows(2, startBtns, stack)

	// Нижний сегмент окна с основной кнопкой Запуска клиента под DirectX
	bottn := container.NewGridWithRows(2, widget.NewButton(data.BTN_START_TITLE, func() { startGameDirectX(&mainWindow) }), manageZone)
	mainWindow.SetContent(container.NewGridWithRows(2, img, bottn))

	mainWindow.Show()
	App.Settings().SetTheme(theme.DarkTheme())
	App.Run()

}

func startUpdate(progress *widget.ProgressBar, progressText *widget.Label, mainWindow *fyne.Window) {
	downloadNewClient(progress, progressText, mainWindow)
	toTemp(progress, progressText, mainWindow)
	replaceTemp(progress, progressText, mainWindow)
	startGameDirectX(mainWindow)
}

func startGameDirectX(w *fyne.Window) {
	currentDir, err := os.Getwd()
	if err != nil {
		showError("16", err, w)
		return
	}

	exePath := filepath.Join(currentDir, data.EXE_NAME_MAIN)

	cmd := exec.Command(exePath)
	err = cmd.Run()
	if err != nil {
		showError("17", err, w)
		return
	}
}

func startGameOpenGl(w *fyne.Window) {
	currentDir, err := os.Getwd()
	if err != nil {
		showError("14", err, w)
		return
	}

	exePath := filepath.Join(currentDir, data.EXE_OPENGL_NAME)

	cmd := exec.Command(exePath)
	err = cmd.Run()
	if err != nil {
		showError("15", err, w)
		return
	}
}

func replaceTemp(progress *widget.ProgressBar, progressText *widget.Label, w *fyne.Window) {
	progress.SetValue(0.8)
	progressText.SetText(data.PROGRESSBAR_TITLE_TOEND)

	dir, err := os.Getwd()
	if err != nil {
		showError("9", err, w)
	}

	tempDir := filepath.Join(dir, "temp")

	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		showError("10", err, w)
	}

	for _, file := range files {
		src := filepath.Join(tempDir, file.Name())
		dst := filepath.Join(dir, file.Name())

		if _, err := os.Stat(dst); err == nil {
			err = os.RemoveAll(dst)
			if err != nil {
				showError("11", err, w)
			}
		}

		err = os.Rename(src, dst)
		if err != nil {
			showError("12", err, w)
		}
	}

	progress.SetValue(0.9)

	err = os.RemoveAll(tempDir)
	if err != nil {
		showError("13", err, w)
	}

	progress.SetValue(1)
	progressText.SetText(data.PROGRESSBAR_TITLE_END)
}

func toTemp(progress *widget.ProgressBar, progressText *widget.Label, w *fyne.Window) {
	progress.SetValue(0.6)
	progressText.SetText(data.PROGRESSBAR_TITLE_UNPACK)

	dir, err := os.Getwd()
	if err != nil {
		showError("4", err, w)
	}

	tempDir := filepath.Join(dir, "temp")
	if err := os.Mkdir(tempDir, 0755); err != nil {
		showError("5", err, w)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} // end if

		if !info.IsDir() && filepath.Ext(path) == ".zip" { //  * if

			if err := unzip(path, tempDir); err != nil { // ** if
				return err
			} // ** end if

		} // * end if

		return nil
	})
	if err != nil {
		showError("6", err, w)
	}

	err = os.Remove(data.NAME_ARCH)
	if err != nil {
		showError("7", err, w)
	}
	err = os.Remove(data.NAME_ARCH_OTHER)
	if err != nil {
		showError("8", err, w)
	}

	progress.SetValue(0.7)
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

func downloadNewClient(progress *widget.ProgressBar, progressText *widget.Label, w *fyne.Window) {
	fileURL := data.FILE_URL
	fileName := data.NAME_ARCH

	progress.SetValue(0)
	progressText.SetText(data.PROGRESSBAR_TITLE_DOWNLOAD)

	response, err := http.Get(fileURL)
	if err != nil {
		showError("1", err, w)
	}
	defer response.Body.Close()

	file, err := os.Create(fileName)
	if err != nil {
		showError("2", err, w)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		showError("3", err, w)
	}

	progress.SetValue(0.5)
}

func showError(errText string, err error, w *fyne.Window) {
	dialog.ShowInformation("Oops", "Ошибка "+errText+"\n"+err.Error(), *w)
}
