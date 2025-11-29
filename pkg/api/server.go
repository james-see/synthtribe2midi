// Package api provides the REST API server for synthtribe2midi
package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/james-see/synthtribe2midi/pkg/converter"
	"github.com/james-see/synthtribe2midi/pkg/converter/devices"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title SynthTribe2MIDI API
// @version 1.0
// @description API for converting between MIDI and Behringer SynthTribe formats
// @host localhost:8080
// @BasePath /api/v1

// StartServer starts the API server on the specified port
func StartServer(port int) error {
	r := gin.Default()
	
	// CORS middleware
	r.Use(corsMiddleware())
	
	// Health check
	r.GET("/health", healthCheck)
	
	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthCheck)
		v1.POST("/convert/midi2seq", handleMIDIToSeq)
		v1.POST("/convert/seq2midi", handleSeqToMIDI)
		v1.POST("/convert/midi2syx", handleMIDIToSyx)
		v1.POST("/convert/syx2midi", handleSyxToMIDI)
		v1.POST("/convert/seq2syx", handleSeqToSyx)
		v1.POST("/convert/syx2seq", handleSyxToSeq)
		v1.GET("/formats", listFormats)
		v1.GET("/devices", listDevices)
	}
	
	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	return r.Run(fmt.Sprintf(":%d", port))
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// healthCheck godoc
// @Summary Health check endpoint
// @Description Returns the health status of the API
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "synthtribe2midi",
	})
}

// listFormats godoc
// @Summary List supported formats
// @Description Returns a list of supported file formats
// @Tags info
// @Produce json
// @Success 200 {object} map[string][]string
// @Router /api/v1/formats [get]
func listFormats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"formats": []string{"midi", "seq", "syx"},
		"conversions": converter.GetSupportedConversions(),
	})
}

// listDevices godoc
// @Summary List supported devices
// @Description Returns a list of supported Behringer devices
// @Tags info
// @Produce json
// @Success 200 {object} map[string][]map[string]string
// @Router /api/v1/devices [get]
func listDevices(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"devices": []map[string]string{
			{"id": "td3", "name": "Behringer TD-3", "description": "TB-303 clone"},
		},
	})
}

// handleMIDIToSeq godoc
// @Summary Convert MIDI to .seq
// @Description Upload a MIDI file and receive a .seq file
// @Tags convert
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param file formance file true "MIDI file to convert"
// @Param device query string false "Target device (default: td3)"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Router /api/v1/convert/midi2seq [post]
func handleMIDIToSeq(c *gin.Context) {
	handleConversion(c, "midi", "seq")
}

// handleSeqToMIDI godoc
// @Summary Convert .seq to MIDI
// @Description Upload a .seq file and receive a MIDI file
// @Tags convert
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param file formance file true ".seq file to convert"
// @Param device query string false "Source device (default: td3)"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Router /api/v1/convert/seq2midi [post]
func handleSeqToMIDI(c *gin.Context) {
	handleConversion(c, "seq", "midi")
}

// handleMIDIToSyx godoc
// @Summary Convert MIDI to .syx
// @Description Upload a MIDI file and receive a .syx file
// @Tags convert
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param file formance file true "MIDI file to convert"
// @Param device query string false "Target device (default: td3)"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Router /api/v1/convert/midi2syx [post]
func handleMIDIToSyx(c *gin.Context) {
	handleConversion(c, "midi", "syx")
}

// handleSyxToMIDI godoc
// @Summary Convert .syx to MIDI
// @Description Upload a .syx file and receive a MIDI file
// @Tags convert
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param file formance file true ".syx file to convert"
// @Param device query string false "Source device (default: td3)"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Router /api/v1/convert/syx2midi [post]
func handleSyxToMIDI(c *gin.Context) {
	handleConversion(c, "syx", "midi")
}

// handleSeqToSyx godoc
// @Summary Convert .seq to .syx
// @Description Upload a .seq file and receive a .syx file
// @Tags convert
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param file formance file true ".seq file to convert"
// @Param device query string false "Device (default: td3)"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Router /api/v1/convert/seq2syx [post]
func handleSeqToSyx(c *gin.Context) {
	handleConversion(c, "seq", "syx")
}

// handleSyxToSeq godoc
// @Summary Convert .syx to .seq
// @Description Upload a .syx file and receive a .seq file
// @Tags convert
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param file formance file true ".syx file to convert"
// @Param device query string false "Device (default: td3)"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Router /api/v1/convert/syx2seq [post]
func handleSyxToSeq(c *gin.Context) {
	handleConversion(c, "syx", "seq")
}

func handleConversion(c *gin.Context, fromFormat, toFormat string) {
	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer func() { _ = file.Close() }()
	
	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file"})
		return
	}
	
	// Get device (default to TD-3)
	deviceName := c.DefaultQuery("device", "td3")
	var device converter.Device
	switch deviceName {
	case "td3", "td-3":
		device = devices.NewTD3()
	default:
		device = devices.NewTD3()
	}
	
	conv := converter.New(device)
	
	// Perform conversion
	var result []byte
	var outputExt string
	
	switch fromFormat + "2" + toFormat {
	case "midi2seq":
		result, err = conv.MIDIToSeq(data)
		outputExt = ".seq"
	case "seq2midi":
		result, err = conv.SeqToMIDI(data)
		outputExt = ".mid"
	case "midi2syx":
		result, err = conv.MIDIToSyx(data)
		outputExt = ".syx"
	case "syx2midi":
		result, err = conv.SyxToMIDI(data)
		outputExt = ".mid"
	case "seq2syx":
		result, err = conv.SeqToSyx(data)
		outputExt = ".syx"
	case "syx2seq":
		result, err = conv.SyxToSeq(data)
		outputExt = ".seq"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported conversion"})
		return
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Generate output filename
	outputName := header.Filename
	if len(outputName) > 4 {
		outputName = outputName[:len(outputName)-4] + outputExt
	} else {
		outputName = "converted" + outputExt
	}
	
	// Set content type and headers
	var contentType string
	switch toFormat {
	case "midi":
		contentType = "audio/midi"
	default:
		contentType = "application/octet-stream"
	}
	
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", outputName))
	c.Data(http.StatusOK, contentType, result)
}

