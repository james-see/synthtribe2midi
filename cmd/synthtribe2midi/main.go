// Package main is the entry point for synthtribe2midi CLI
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/james-see/synthtribe2midi/pkg/api"
	"github.com/james-see/synthtribe2midi/pkg/converter"
	"github.com/james-see/synthtribe2midi/pkg/converter/devices"
	"github.com/james-see/synthtribe2midi/pkg/tui"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	outputFile string
	deviceName string
	serverPort int
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "synthtribe2midi",
	Short: "Convert between MIDI and Behringer SynthTribe formats",
	Long: `synthtribe2midi is a tool for converting between standard MIDI files 
and Behringer SynthTribe .seq/.syx formats.

Supports TD-3 (TB-303 clone) patterns with extensibility for other devices.

Examples:
  synthtribe2midi convert pattern.mid -o pattern.seq
  synthtribe2midi midi2seq pattern.mid -o pattern.seq
  synthtribe2midi seq2midi pattern.seq -o pattern.mid
  synthtribe2midi tui
  synthtribe2midi serve --port 8080`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

var convertCmd = &cobra.Command{
	Use:   "convert <input>",
	Short: "Auto-detect and convert between formats",
	Long:  `Automatically detects input format and converts to the output format based on file extension.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConvert,
}

var midi2seqCmd = &cobra.Command{
	Use:   "midi2seq <input.mid>",
	Short: "Convert MIDI to .seq format",
	Args:  cobra.ExactArgs(1),
	RunE:  runMIDIToSeq,
}

var seq2midiCmd = &cobra.Command{
	Use:   "seq2midi <input.seq>",
	Short: "Convert .seq to MIDI format",
	Args:  cobra.ExactArgs(1),
	RunE:  runSeqToMIDI,
}

var midi2syxCmd = &cobra.Command{
	Use:   "midi2syx <input.mid>",
	Short: "Convert MIDI to .syx format",
	Args:  cobra.ExactArgs(1),
	RunE:  runMIDIToSyx,
}

var syx2midiCmd = &cobra.Command{
	Use:   "syx2midi <input.syx>",
	Short: "Convert .syx to MIDI format",
	Args:  cobra.ExactArgs(1),
	RunE:  runSyxToMIDI,
}

var seq2syxCmd = &cobra.Command{
	Use:   "seq2syx <input.seq>",
	Short: "Convert .seq to .syx format",
	Args:  cobra.ExactArgs(1),
	RunE:  runSeqToSyx,
}

var syx2seqCmd = &cobra.Command{
	Use:   "syx2seq <input.syx>",
	Short: "Convert .syx to .seq format",
	Args:  cobra.ExactArgs(1),
	RunE:  runSyxToSeq,
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI",
	RunE:  runTUI,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	RunE:  runServe,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&deviceName, "device", "d", "td3", "Target device (td3)")

	// Convert command
	convertCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (required)")
	_ = convertCmd.MarkFlagRequired("output")

	// midi2seq command
	midi2seqCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output .seq file path")

	// seq2midi command
	seq2midiCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output .mid file path")

	// midi2syx command
	midi2syxCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output .syx file path")

	// syx2midi command
	syx2midiCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output .mid file path")

	// seq2syx command
	seq2syxCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output .syx file path")

	// syx2seq command
	syx2seqCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output .seq file path")

	// serve command
	serveCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Server port")

	// Add commands
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(midi2seqCmd)
	rootCmd.AddCommand(seq2midiCmd)
	rootCmd.AddCommand(midi2syxCmd)
	rootCmd.AddCommand(syx2midiCmd)
	rootCmd.AddCommand(seq2syxCmd)
	rootCmd.AddCommand(syx2seqCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(serveCmd)
}

func getDevice() converter.Device {
	switch strings.ToLower(deviceName) {
	case "td3", "td-3":
		return devices.NewTD3()
	default:
		return devices.NewTD3()
	}
}

func getOutputPath(input, defaultExt string) string {
	if outputFile != "" {
		return outputFile
	}
	base := strings.TrimSuffix(input, filepath.Ext(input))
	return base + defaultExt
}

func runConvert(cmd *cobra.Command, args []string) error {
	input := args[0]
	conv := converter.New(getDevice())
	
	fmt.Printf("Converting %s -> %s\n", input, outputFile)
	if err := conv.ConvertFile(input, outputFile); err != nil {
		return err
	}
	fmt.Println("Conversion complete!")
	return nil
}

func runMIDIToSeq(cmd *cobra.Command, args []string) error {
	input := args[0]
	output := getOutputPath(input, ".seq")
	
	conv := converter.New(getDevice())
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	
	result, err := conv.MIDIToSeq(data)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(output, result, 0644); err != nil {
		return err
	}
	
	fmt.Printf("Converted %s -> %s\n", input, output)
	return nil
}

func runSeqToMIDI(cmd *cobra.Command, args []string) error {
	input := args[0]
	output := getOutputPath(input, ".mid")
	
	conv := converter.New(getDevice())
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	
	result, err := conv.SeqToMIDI(data)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(output, result, 0644); err != nil {
		return err
	}
	
	fmt.Printf("Converted %s -> %s\n", input, output)
	return nil
}

func runMIDIToSyx(cmd *cobra.Command, args []string) error {
	input := args[0]
	output := getOutputPath(input, ".syx")
	
	conv := converter.New(getDevice())
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	
	result, err := conv.MIDIToSyx(data)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(output, result, 0644); err != nil {
		return err
	}
	
	fmt.Printf("Converted %s -> %s\n", input, output)
	return nil
}

func runSyxToMIDI(cmd *cobra.Command, args []string) error {
	input := args[0]
	output := getOutputPath(input, ".mid")
	
	conv := converter.New(getDevice())
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	
	result, err := conv.SyxToMIDI(data)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(output, result, 0644); err != nil {
		return err
	}
	
	fmt.Printf("Converted %s -> %s\n", input, output)
	return nil
}

func runSeqToSyx(cmd *cobra.Command, args []string) error {
	input := args[0]
	output := getOutputPath(input, ".syx")
	
	conv := converter.New(getDevice())
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	
	result, err := conv.SeqToSyx(data)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(output, result, 0644); err != nil {
		return err
	}
	
	fmt.Printf("Converted %s -> %s\n", input, output)
	return nil
}

func runSyxToSeq(cmd *cobra.Command, args []string) error {
	input := args[0]
	output := getOutputPath(input, ".seq")
	
	conv := converter.New(getDevice())
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	
	result, err := conv.SyxToSeq(data)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(output, result, 0644); err != nil {
		return err
	}
	
	fmt.Printf("Converted %s -> %s\n", input, output)
	return nil
}

func runTUI(cmd *cobra.Command, args []string) error {
	return tui.Run()
}

func runServe(cmd *cobra.Command, args []string) error {
	fmt.Printf("Starting API server on port %d...\n", serverPort)
	return api.StartServer(serverPort)
}

