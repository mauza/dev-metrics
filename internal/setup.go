package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	modelSize string
	modelsDir string
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Download and setup the LLaMA model",
	RunE:  runSetup,
}

func init() {
	setupCmd.Flags().StringVar(&modelSize, "model-size", "small",
		"Model size to download (tiny, small, medium)")
	setupCmd.Flags().StringVar(&modelsDir, "models-dir", "./models",
		"Directory to store models")
}

type modelInfo struct {
	Name        string
	URL         string
	Size        string
	Description string
}

var models = map[string]modelInfo{
	"tiny": {
		Name:        "llama-2-7b-chat.Q2_K.gguf",
		URL:         "https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGUF/resolve/main/llama-2-7b-chat.Q2_K.gguf",
		Size:        "2.87GB",
		Description: "Smallest, fastest, lowest quality",
	},
	"small": {
		Name:        "llama-2-7b-chat.Q4_K_M.gguf",
		URL:         "https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGUF/resolve/main/llama-2-7b-chat.Q4_K_M.gguf",
		Size:        "4.31GB",
		Description: "Good balance of speed and quality",
	},
	"medium": {
		Name:        "llama-2-7b-chat.Q5_K_M.gguf",
		URL:         "https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGUF/resolve/main/llama-2-7b-chat.Q5_K_M.gguf",
		Size:        "5.04GB",
		Description: "Better quality, slower",
	},
}

func runSetup(cmd *cobra.Command, args []string) error {
	model, ok := models[modelSize]
	if !ok {
		return fmt.Errorf("invalid model size: %s", modelSize)
	}

	// Create models directory
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	// Create .gitignore if it doesn't exist
	gitignorePath := filepath.Join(modelsDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte("*.gguf\n"), 0644); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
	}

	fmt.Printf("\nSelected model: %s\n", model.Name)
	fmt.Printf("Size: %s\n", model.Size)
	fmt.Printf("Description: %s\n", model.Description)

	modelPath := filepath.Join(modelsDir, model.Name)
	if _, err := os.Stat(modelPath); err == nil {
		fmt.Printf("\nModel already exists at %s\n", modelPath)
		fmt.Print("Do you want to download it again? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return nil
		}
		if err := os.Remove(modelPath); err != nil {
			return fmt.Errorf("failed to remove existing model: %w", err)
		}
	}

	fmt.Printf("\nDownloading model to %s\n", modelPath)
	if err := downloadFile(model.URL, modelPath); err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}

	fmt.Println("\nSetup complete! You can now use dev-metrics with the downloaded model.")
	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create progress bar
	total := resp.ContentLength
	current := int64(0)
	lastPercent := 0

	reader := io.TeeReader(resp.Body, &progressWriter{
		total:      total,
		current:    &current,
		lastOutput: &lastPercent,
	})

	_, err = io.Copy(out, reader)
	fmt.Println() // New line after progress bar
	return err
}

type progressWriter struct {
	total      int64
	current    *int64
	lastOutput *int
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	*pw.current += int64(n)
	percent := int(float64(*pw.current) / float64(pw.total) * 100)

	if percent > *pw.lastOutput {
		fmt.Printf("\rDownloading... %d%%", percent)
		*pw.lastOutput = percent
	}

	return n, nil
}
