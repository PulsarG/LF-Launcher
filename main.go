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
	"image/color"
	"io"
	"io/fs"
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
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()
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

func startUpdate(progressBar *widget.ProgressBarInfinite, progressText *widget.Label, mainWindow *fyne.Window) {
	wtColor := color.RGBA{255, 255, 255, 255}
	title := canvas.NewText("Для успешного обновления необходимо минимум 2 ГБ свободного места", wtColor)
	dialog.ShowCustomConfirm("Внимание", "Обновить", "Отмена", title, func(b bool) {
		if b {
			downloadNewClient(progressBar, progressText, mainWindow)
			toTemp(progressText, mainWindow)
			replaceTemp(progressText, mainWindow)
			progressBar.Hide()

			// Автостарт после обновления
			// startGameDirectX(mainWindow)
		} // end if
	}, *mainWindow) // end dialog
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

func replaceTemp(progressText *widget.Label, w *fyne.Window) {
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

	// Т.к. распаковка архива вынимает ВСЕ файлы
	// то необходимо перед перемещением нужных файлов (архивы в директории Resources)
	// удалить те, что не нужны
	cleaningTemp(dir, tempDir, &files, w)

	// Перемещаем новые файлы клиента из temp
	for _, file := range files {
		src := filepath.Join(tempDir, file.Name())
		dst := filepath.Join(dir, file.Name())

		if _, err := os.Stat(dst); err == nil {
			err = os.RemoveAll(dst)
			if err != nil {
				showError("11", err, w)
			}
		} // end if

		err = os.Rename(src, dst)
		if err != nil {
			continue
		}
	} // end for

	err = os.RemoveAll(tempDir)
	if err != nil {
		showError("13", err, w)
	}

	progressText.SetText(data.PROGRESSBAR_TITLE_END)
}

func cleaningTemp(dir string, tempDir string, files *[]fs.FileInfo, w *fyne.Window) {
	for _, file := range *files { // for
		path := fmt.Sprintf("%s/%s", tempDir, file.Name())

		if (file.Name() != data.EXE_OPENGL_NAME) && (file.Name() != data.EXE_NAME_MAIN) && (file.Name() != data.FILE_TEXT) && (file.Name() != data.RESOURCES_DIR_NAME) {
			err := os.RemoveAll(path)
			if err != nil { // ** if
				showError("19", err, w)
			} // ** end if
		} // * end if
	} // end for
}

func toTemp(progressText *widget.Label, w *fyne.Window) {
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

func downloadNewClient(progressBar *widget.ProgressBarInfinite, progressText *widget.Label, w *fyne.Window) {
	progressBar.Show()
	progressText.SetText(data.PROGRESSBAR_TITLE_DOWNLOAD)

	response, err := http.Get(data.FILE_URL)
	if err != nil {
		showError("1", err, w)
	}
	defer response.Body.Close()

	file, err := os.Create(data.NAME_ARCH)
	if err != nil {
		showError("2", err, w)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		showError("3", err, w)
	}

}

func showError(errText string, err error, w *fyne.Window) {
	dialog.ShowInformation("Oops", "Ошибка "+errText+"\n"+err.Error(), *w)
}
