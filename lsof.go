//
// lsof list opened file+pid
// support linux only
//

package lsof

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// File2PIDs list PIDs who opened this file
type File2PIDs struct {
	File string           // the file(full path)
	PIDs map[int]struct{} // file opened by PIDs
}

//
func newFile2PIDs(f string) *File2PIDs {
	fdi := new(File2PIDs)
	fdi.PIDs = make(map[int]struct{})
	fdi.File = f
	return fdi
}

// PID2Files list files opened by this PID
type PID2Files struct {
	PID   int                 // file opened by PIDs
	Files map[string]struct{} // the file(full path)
}

//
func newPID2Files(pid int) *PID2Files {
	pidi := new(PID2Files)
	pidi.Files = make(map[string]struct{})
	pidi.PID = pid
	return pidi
}

// InfoList list all opened files
type InfoList struct {
	files    map[string]*File2PIDs // list index by file
	pids     map[int]*PID2Files    // list index by pid
	readDone chan struct{}         // lock for reading done
}

// newInfoList
func newInfoList() *InfoList {
	l := new(InfoList)
	l.files = make(map[string]*File2PIDs)
	l.pids = make(map[int]*PID2Files)
	l.readDone = make(chan struct{})
	return l
}

// Open start a goroutine and read
func Open(listdir string, prefix string) (*InfoList, error) {
	var err error
	if len(listdir) == 0 {
		listdir = "/proc/"
	}
	if prefix == "" {
		prefix = "."
	}
	prefix = filepath.Clean(prefix)
	listdir, err = filepath.Abs(listdir)
	if err != nil {
		return nil, err
	}
	l := newInfoList()
	//fmt.Printf("Open: reading %s ...\n", listdir)
	err = l.readDir(listdir, prefix)
	if err != nil {
		fmt.Printf("readDir %s: %s\n", listdir, err.Error())
	}
	close(l.readDone)
	return l, err
}

// Lsof start a goroutine and read File2PIDs send to output channel
func Lsof(prefix string) (*InfoList, error) {
	return Open("", prefix)
}

// LsofPID list opened file of pid
func LsofPID(pid int, prefix string) (*InfoList, error) {
	l := newInfoList()
	fddir := "/proc/" + fmt.Sprintf("%d", pid) + "/fd/"
	//fmt.Printf("Open: reading %s ...\n", fddir)
	err := l.readPIDDir(pid, fddir, prefix)
	if err != nil {
		fmt.Printf("LsofPID %s: %s\n", err.Error())
	}
	close(l.readDone)
	return l, err
}

// readPIDDir
func (l *InfoList) readPIDDir(onepid int, fddir string, prefix string) error {
	var err error
	var symlist []string
	var onefile string
	var checkprefix bool = true
	if prefix == "." {
		checkprefix = false
	}
	symlist, err = filepath.Glob(fddir)
	if err != nil {
		//fmt.Printf("Error: Glob %s: %s\n", fddir, err.Error())
		return err
	}
	if len(symlist) == 0 {
		return nil
	}
	for _, onefd := range symlist {
		onefile, err = filepath.EvalSymlinks(onefd)
		if err != nil {
			//fmt.Printf("Error: EvalSymlinks %s: %s\n", onefd, err.Error())
			continue
		}
		if checkprefix {
			if strings.HasPrefix(onefile, prefix) == false {
				//fmt.Printf("Skip: pid %d open fd %s -> %s\n", onepid, onefd, onefile)
				continue
			}
		}
		//fmt.Printf("Got: pid %d open fd %s -> %s\n", onepid, onefd, onefile)
		// get onefile from sym link
		if _, ok := l.pids[onepid]; !ok {
			l.pids[onepid] = newPID2Files(onepid)
		}
		l.pids[onepid].Files[onefile] = struct{}{}
		if _, ok := l.files[onefile]; !ok {
			l.files[onefile] = newFile2PIDs(onefile)
		}
		l.files[onefile].PIDs[onepid] = struct{}{}
	}
	return nil
}

// readDir
func (l *InfoList) readDir(path string, prefix string) error {
	var err error
	var cwd string

	cwd, err = os.Getwd()
	if err != nil {
		fmt.Printf("Getwd: %s\n", err.Error())
		cwd = "/"
	}
	defer func() {
		err := os.Chdir(cwd)
		if err != nil {
			fmt.Printf("WARNING: Chdir %s: %s\n", cwd, err.Error())
		}
	}()
	err = os.Chdir(path)
	if err != nil {
		fmt.Printf("Error: Chdir %s: %s\n", path, err.Error())
		return err
	}
	// list pid dir in current dir
	var pidlist []string
	pidlist, err = filepath.Glob("*")
	if err != nil {
		fmt.Printf("Error: Glob %s: %s\n", path, err.Error())
		return err
	}
	for _, val := range pidlist {
		onepid, err := strconv.Atoi(val)
		if err != nil {
			// no a pid-dir
			//fmt.Printf("Error: Atoi %s: %s\n", path+"/"+val, err.Error())
			continue
		}
		fddir := path + "/" + val + "/fd/*"
		//func (l *InfoList) readPIDDir(onepid int, fddir string, prefix string) error
		err = l.readPIDDir(onepid, fddir, prefix)
		if err != nil {
			fmt.Printf("Error: readPIDDir %s: %s\n", fddir, err.Error())
			continue
		}
	}
	return nil
}

// sendFile2PIDs send File2PIDs list to channel
func (l *InfoList) sendFile2PIDs(lc chan *File2PIDs) {
	for _, val := range l.files {
		lc <- val
	}
	close(lc)
	return
}

// sendPID2Files send PID2Files list to channel
func (l *InfoList) sendPID2Files(lc chan *PID2Files) {
	for _, val := range l.pids {
		lc <- val
	}
	close(lc)
	return
}

// File2PIDsMap return file list in map
func (l *InfoList) File2PIDsMap() map[string]*File2PIDs {
	// block until read done
	//fmt.Printf("File2PIDsMap: waiting for read ...\n")
	<-l.readDone
	//fmt.Printf("File2PIDsMap: read done ...\n")
	return l.files
}

// File2PIDsChan return file list in chan
func (l *InfoList) File2PIDsChan() chan *File2PIDs {
	// block until read done
	//fmt.Printf("File2PIDsChan: waiting for read ...\n")
	<-l.readDone
	//fmt.Printf("File2PIDsChan: read done ...\n")
	lc := make(chan *File2PIDs, 1024)
	go l.sendFile2PIDs(lc)
	return lc
}

// PID2FilesMap return file list in map
func (l *InfoList) PID2FilesMap() map[int]*PID2Files {
	// block until read done
	//fmt.Printf("PID2FilesMap: waiting for read ...\n")
	<-l.readDone
	//fmt.Printf("PID2FilesMap: read done ...\n")
	return l.pids
}

// PID2FilesChan return file list in chan
func (l *InfoList) PID2FilesChan() chan *PID2Files {
	// block until read done
	//fmt.Printf("PID2FilesChan: waiting for read ...\n")
	<-l.readDone
	//fmt.Printf("PID2FilesChan: read done ...\n")
	lc := make(chan *PID2Files, 1024)
	go l.sendPID2Files(lc)
	return lc
}

//
