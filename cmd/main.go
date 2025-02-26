package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

type Stack struct {
	StackName string `json:"StackName"`
}

func main() {
	app := tview.NewApplication().EnableMouse(true) // Habilitar soporte para mouse

	info := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	configFilePath := ""
	inputField := tview.NewInputField().
		SetLabel("Config path: ").
		SetFieldWidth(40).
		SetChangedFunc(func(text string) {
			configFilePath = text
		})

	templates := tview.NewList().ShowSecondaryText(false)
	templates.SetBorder(true).SetTitle("Templates")

	accountDropdown := tview.NewDropDown().
		SetLabel("Seleccione una cuenta de AWS: ").
		SetFieldWidth(20)

	scanButton := tview.NewButton("Scan").
		SetSelectedFunc(func() {
			_, account := accountDropdown.GetCurrentOption()
			if account == "" {
				info.SetText("Debes seleccionar una cuenta.")
				return
			}
			info.SetText(fmt.Sprintf("Escaneando stacks en la cuenta: %s", account))
			output, err := executeShellCommand("aws", "cloudformation", "list-stacks", "--output", "json", "--profile", account)
			if err != nil {
				info.SetText(fmt.Sprintf("Error ejecutando comando: %v\nSalida: %s", err, output))
				return
			}
			var result map[string][]Stack
			if err := json.Unmarshal(output, &result); err != nil {
				info.SetText(fmt.Sprintf("Error parseando JSON: %v", err))
				return
			}
			templates.Clear()
			seenTemplates := make(map[string]bool)
			for _, stack := range result["StackSummaries"] {
				if !seenTemplates[stack.StackName] {
					templates.AddItem(stack.StackName, "", 0, nil)
					seenTemplates[stack.StackName] = true
				}
			}
		})

	loadConfigButton := tview.NewButton("Cargar configuración").
		SetSelectedFunc(func() {
			if configFilePath == "" {
				info.SetText("Debes ingresar la ruta al archivo de configuración.")
				return
			}
			configFile, err := os.Open(configFilePath)
			if err != nil {
				info.SetText(fmt.Sprintf("No se pudo abrir el archivo de configuración: %v", err))
				return
			}
			defer configFile.Close()

			scanner := bufio.NewScanner(configFile)
			var accounts []string
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "[profile ") && strings.HasSuffix(line, "]") {
					profileName := strings.TrimSuffix(strings.TrimPrefix(line, "[profile "), "]")
					accounts = append(accounts, profileName)
				}
			}
			if err := scanner.Err(); err != nil {
				info.SetText(fmt.Sprintf("Error al leer el archivo de configuración: %v", err))
				return
			}

			accountDropdown.SetOptions(accounts, func(option string, index int) {
				info.SetText(fmt.Sprintf("Seleccionaste la cuenta: %s", option))
			})
			info.SetText("Archivo de configuración cargado exitosamente.")
		})

	useCliAccountCheckbox := tview.NewCheckbox().
		SetLabel("Usar cuenta configurada en la CLI: ").
		SetChangedFunc(func(checked bool) {
			if checked {
				accountDropdown.SetDisabled(true)
				inputField.SetDisabled(true)
				loadConfigButton.SetDisabled(true)
				info.SetText("Usando cuenta configurada en la CLI")
				output, err := executeShellCommand("aws", "cloudformation", "list-stacks", "--output", "json")
				if err != nil {
					info.SetText(fmt.Sprintf("Error ejecutando comando: %v\nSalida: %s", err, output))
					return
				}
				var result map[string][]Stack
				if err := json.Unmarshal(output, &result); err != nil {
					info.SetText(fmt.Sprintf("Error parseando JSON: %v", err))
					return
				}
				templates.Clear()
				seenTemplates := make(map[string]bool)
				for _, stack := range result["StackSummaries"] {
					if !seenTemplates[stack.StackName] {
						templates.AddItem(stack.StackName, "", 0, nil)
						seenTemplates[stack.StackName] = true
					}
				}
			} else {
				accountDropdown.SetDisabled(false)
				inputField.SetDisabled(false)
				loadConfigButton.SetDisabled(false)
			}
		})
	templatePathInput := tview.NewInputField().
		SetLabel("Ruta del template actualizado: ").
		SetFieldWidth(40).
		SetDisabled(true)

	templateNameInput := tview.NewInputField().
		SetLabel("Nombre del template: ").
		SetFieldWidth(40).
		SetDisabled(true)

	// Añadir un nuevo campo de entrada para la ruta del archivo de parámetros
	parametersPathInput := tview.NewInputField().
		SetLabel("Ruta del archivo de parámetros: ").
		SetFieldWidth(40).
		SetDisabled(true)

	// Añadir un nuevo checkbox para los parámetros
	parametersCheckbox := tview.NewCheckbox().
		SetLabel("Usar archivo de parámetros: ").
		SetChangedFunc(func(checked bool) {
			if checked {
				parametersPathInput.SetDisabled(false)
			} else {
				parametersPathInput.SetDisabled(true)
			}
		})

	changeSetTypeDropdown := tview.NewDropDown().
		SetLabel("Tipo de Change Set: ").
		SetOptions([]string{"CREATE", "UPDATE"}, nil).
		SetDisabled(true)

	actionDropdown := tview.NewDropDown().
		SetLabel("Seleccione una acción: ").
		SetOptions([]string{"Deploy", "Create-Change-Set"}, func(option string, index int) {
			info.SetText(fmt.Sprintf("Seleccionaste la acción: %s", option))
			if option == "Create-Change-Set" {
				templatePathInput.SetDisabled(false)
				templateNameInput.SetDisabled(false)
				parametersCheckbox.SetDisabled(false)
				changeSetTypeDropdown.SetDisabled(false)
			} else {
				templatePathInput.SetDisabled(true)
				templateNameInput.SetDisabled(true)
				parametersCheckbox.SetDisabled(true)
				parametersPathInput.SetDisabled(true)
				changeSetTypeDropdown.SetDisabled(true)
			}
		})

	executeButton := tview.NewButton("Ejecutar").
		SetSelectedFunc(func() {
			_, account := accountDropdown.GetCurrentOption()
			_, action := actionDropdown.GetCurrentOption()
			templateName := templateNameInput.(*tview.InputField).GetText()
			templatePath := templatePathInput.(*tview.InputField).GetText()
			parametersPath := parametersPathInput.(*tview.InputField).GetText()
			_, changeSetType := changeSetTypeDropdown.(*tview.DropDown).GetCurrentOption()
			if account == "" || action == "" {
				info.SetText("Debes seleccionar una cuenta, un servicio y una acción.")
				return
			}
			if (action == "Deploy" || action == "Create-Change-Set") && (templateName == "" || templatePath == "") {
				info.SetText("Debes ingresar el nombre del template y proporcionar la ruta del template actualizado.")
				return
			}
			if action == "Create-Change-Set" && parametersCheckbox.IsChecked() && parametersPath == "" {
				info.SetText("Debes proporcionar la ruta del archivo de parámetros.")
				return
			}
			info.SetText(fmt.Sprintf("Ejecutando acción: %s en la cuenta: %s", action, account))
			var output []byte
			var err error
			if action == "Create-Change-Set" {
				changeSetName := "aws-explorer-" + templateName
				if useCliAccountCheckbox.IsChecked() {
					output, err = executeShellCommand("aws", "cloudformation", "create-change-set", "--change-set-name", changeSetName, "--stack-name", "infra-alarms", "--change-set-type", changeSetType, "--template-body", "file://"+templatePath, "--parameters", "file://"+parametersPath)
				} else {
					output, err = executeShellCommand("aws", "cloudformation", "create-change-set", "--change-set-name", changeSetName, "--stack-name", "infra-alarms", "--change-set-type", changeSetType, "--template-body", "file://"+templatePath, "--parameters", "file://"+parametersPath, "--profile", account)
				}
			} else {
				if useCliAccountCheckbox.IsChecked() {
					output, err = executeShellCommand("aws", "cloudformation", action, "--stack-name", templateName, "--template-file", templatePath)
				} else {
					output, err = executeShellCommand("aws", "cloudformation", action, "--stack-name", templateName, "--template-file", templatePath, "--profile", account)
				}
			}
			if err != nil {
				info.SetText(fmt.Sprintf("Error ejecutando comando: %v\nSalida: %s", err, output))
				return
			}
			info.SetText(fmt.Sprintf("Acción %s ejecutada exitosamente en la cuenta %s", action, account))
		})

	exitButton := tview.NewButton("Exit").
		SetSelectedFunc(func() {
			app.Stop()
		})

	// Crear un Flex layout para los botones
	buttonFlex := tview.NewFlex().
		AddItem(exitButton, 0, 1, false).   // Botón "Exit" a la izquierda
		AddItem(nil, 0, 1, false).          // Espacio vacío en el medio
		AddItem(executeButton, 0, 1, false) // Botón "Scan" a la derecha

	// Crear un Flex layout para las tablas y botones
	tablesFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(templates, 0, 1, false).
		AddItem(buttonFlex, 1, 1, false) // Ajustar el tamaño del buttonFlex

	// Crear el layout principal
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
						AddItem(tablesFlex, 0, 1, false).
						AddItem(tview.NewFlex().
							SetDirection(tview.FlexRow).
							AddItem(nil, 5, 1, false). // Espacio vacío en el medio
							AddItem(inputField, 3, 1, true).
							AddItem(loadConfigButton, 1, 0, false).
							AddItem(nil, 1, 0, false).
							AddItem(accountDropdown, 3, 1, true).
							AddItem(actionDropdown, 3, 1, true).
							AddItem(templatePathInput, 3, 1, true).
							AddItem(templateNameInput, 3, 1, true).
							AddItem(useCliAccountCheckbox, 3, 1, true).
							AddItem(parametersCheckbox, 3, 1, true).
							AddItem(parametersPathInput, 3, 1, true).
							AddItem(changeSetTypeDropdown, 3, 1, true). // Añadir el nuevo dropdown
							AddItem(scanButton, 1, 0, false), 0, 1, true), 0, 1, true).
		AddItem(nil, 3, 1, false). // Espacio vacío en el medio
		AddItem(info, 10, 3, false)

	// Configurar la aplicación
	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}
}

func executeShellCommand(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error ejecutando comando: %v\nSalida: %s", err, output)
	}
	return output, nil
}
