package log

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/spf13/viper"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FileLogDriver struct {
	// settings
	logFileName string
	rootPath    string

	// internals
	mu sync.Mutex
	interfaces.Logger
}

func (d *FileLogDriver) Init() {
	go d.cleanup()
}

func (d *FileLogDriver) Close() (err error) {
	return nil
}

func (d *FileLogDriver) WriteLine(id string, line string) (err error) {
	d.initDir(id)

	d.mu.Lock()
	defer d.mu.Unlock()
	filePath := d.getLogFilePath(id, d.logFileName)

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(0760))
	if err != nil {
		d.Errorf("open file error: %v", err)
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			d.Errorf("close file error: %v", err)
			return
		}
	}(f)

	_, err = f.WriteString(line + "\n")
	if err != nil {
		d.Errorf("write file error: %v", err)
		return err
	}

	return nil
}

func (d *FileLogDriver) WriteLines(id string, lines []string) (err error) {
	linesString := strings.Join(lines, "\n")
	if err := d.WriteLine(id, linesString); err != nil {
		d.Errorf("write line error: %v", err)
		return err
	}
	return nil
}

func (d *FileLogDriver) Find(id string, pattern string, skip int, limit int) (lines []string, err error) {
	if pattern != "" {
		err = fmt.Errorf("find with pattern not implemented")
		d.Errorf("%v", err)
		return lines, err
	}
	if !utils.Exists(d.getLogFilePath(id, d.logFileName)) {
		return nil, nil
	}

	f, err := os.Open(d.getLogFilePath(id, d.logFileName))
	if err != nil {
		d.Errorf("failed to open file: %v", err)
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewReaderSize(f, 1024*1024*10)

	i := -1
	for {
		line, err := sc.ReadString(byte('\n'))
		if err != nil {
			break
		}
		line = strings.TrimSuffix(line, "\n")

		i++

		if i < skip {
			continue
		}

		if i >= skip+limit {
			break
		}

		lines = append(lines, line)
	}

	return lines, nil
}

func (d *FileLogDriver) Count(id string, pattern string) (n int, err error) {
	if pattern != "" {
		err = fmt.Errorf("count with pattern not implemented")
		d.Errorf("%v", err)
		return n, err
	}
	if !utils.Exists(d.getLogFilePath(id, d.logFileName)) {
		return 0, nil
	}

	f, err := os.Open(d.getLogFilePath(id, d.logFileName))
	if err != nil {
		d.Errorf("failed to open file: %v", err)
		return n, err
	}
	return d.lineCounter(f)
}

func (d *FileLogDriver) Flush() (err error) {
	return nil
}

func (d *FileLogDriver) getBasePath(id string) (filePath string) {
	return filepath.Join(utils.GetTaskLogPath(), id)
}

func (d *FileLogDriver) getMetadataPath(id string) (filePath string) {
	return filepath.Join(d.getBasePath(id), MetadataName)
}

func (d *FileLogDriver) getLogFilePath(id, fileName string) (filePath string) {
	return filepath.Join(d.getBasePath(id), fileName)
}

func (d *FileLogDriver) getLogFiles(id string) (files []os.FileInfo) {
	files, err := utils.ListDir(d.getBasePath(id))
	if err != nil {
		d.Errorf("failed to list log files: %v", err)
		return nil
	}
	return
}

func (d *FileLogDriver) initDir(id string) {
	if !utils.Exists(d.getBasePath(id)) {
		if err := os.MkdirAll(d.getBasePath(id), os.FileMode(0770)); err != nil {
			d.Errorf("failed to create log directory: %s", d.getBasePath(id))
			return
		}
	}
}

func (d *FileLogDriver) lineCounter(r io.Reader) (n int, err error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

func (d *FileLogDriver) getTtl() time.Duration {
	ttl := viper.GetString("log.ttl")
	if ttl == "" {
		return DefaultLogTtl
	}

	if strings.HasSuffix(ttl, "s") {
		ttl = strings.TrimSuffix(ttl, "s")
		n, err := strconv.Atoi(ttl)
		if err != nil {
			return DefaultLogTtl
		}
		return time.Duration(n) * time.Second
	} else if strings.HasSuffix(ttl, "m") {
		ttl = strings.TrimSuffix(ttl, "m")
		n, err := strconv.Atoi(ttl)
		if err != nil {
			return DefaultLogTtl
		}
		return time.Duration(n) * time.Minute
	} else if strings.HasSuffix(ttl, "h") {
		ttl = strings.TrimSuffix(ttl, "h")
		n, err := strconv.Atoi(ttl)
		if err != nil {
			return DefaultLogTtl
		}
		return time.Duration(n) * time.Hour

	} else if strings.HasSuffix(ttl, "d") {
		ttl = strings.TrimSuffix(ttl, "d")
		n, err := strconv.Atoi(ttl)
		if err != nil {
			return DefaultLogTtl
		}
		return time.Duration(n) * 24 * time.Hour
	} else {
		return DefaultLogTtl
	}
}

func (d *FileLogDriver) cleanup() {
	// check if log path is set
	if utils.GetTaskLogPath() == "" {
		d.Errorf("log path is not set")
		return
	}

	// check if log path exists
	if !utils.Exists(utils.GetTaskLogPath()) {
		// create log directory if not exists
		if err := os.MkdirAll(utils.GetTaskLogPath(), os.FileMode(0770)); err != nil {
			d.Errorf("failed to create log directory: %s", utils.GetTaskLogPath())
			return
		}
	}

	ticker := time.NewTicker(10 * time.Minute)

	for {
		select {
		case <-ticker.C:
			dirs, err := utils.ListDir(utils.GetTaskLogPath())
			if err != nil {
				d.Errorf("failed to list log directory: %s", utils.GetTaskLogPath())
				continue
			}
			for _, dir := range dirs {
				if time.Now().After(dir.ModTime().Add(d.getTtl())) {
					if err := os.RemoveAll(d.getBasePath(dir.Name())); err != nil {
						d.Errorf("failed to remove outdated log directory: %s", d.getBasePath(dir.Name()))
						continue
					}
					d.Infof("removed outdated log directory: %s", d.getBasePath(dir.Name()))
				}
			}
		}
	}
}

func newFileLogDriver() Driver {
	// driver
	driver := &FileLogDriver{
		logFileName: "log.txt",
		mu:          sync.Mutex{},
		Logger:      utils.NewLogger("FileLogDriver"),
	}

	// init
	driver.Init()

	return driver
}

var logDriver Driver
var logDriverOnce sync.Once

func GetFileLogDriver() Driver {
	logDriverOnce.Do(func() {
		logDriver = newFileLogDriver()
	})
	return logDriver
}
