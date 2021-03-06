// +-------------------------------------------------------------------------
// | Copyright (C) 2017 Yunify, Inc.
// +-------------------------------------------------------------------------
// | Licensed under the Apache License, Version 2.0 (the "License");
// | you may not use this work except in compliance with the License.
// | You may obtain a copy of the License in the LICENSE file, or at:
// |
// | http://www.apache.org/licenses/LICENSE-2.0
// |
// | Unless required by applicable law or agreed to in writing, software
// | distributed under the License is distributed on an "AS IS" BASIS,
// | WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// | See the License for the specific language governing permissions and
// | limitations under the License.
// +-------------------------------------------------------------------------

package client

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

func (c *Handler) absPath(p string) string {
	p2 := c.Path()

	if strings.HasPrefix(p, "/") {
		p2 = p
	} else {
		if p2 != "/" {
			p2 += "/"
		}
		p2 += p
	}

	if p2 != "/" && strings.HasSuffix(p2, "/") {
		p2 = p2[0 : len(p2)-1]
	}

	return p2
}

func (c *Handler) handleCWD() {
	if c.param == ".." {
		c.handleCDUP()
		return
	}

	p := c.absPath(c.param)

	if err := c.driver.ChangeDirectory(c, p); err == nil {
		c.SetPath(p)
		c.WriteMessage(250, fmt.Sprintf("CD worked on %s", p))
	} else {
		c.WriteMessage(550, fmt.Sprintf("CD issue: %v", err))
	}
}

func (c *Handler) handleMKD() {
	p := c.absPath(c.param)
	if err := c.driver.MakeDirectory(c, p); err == nil {
		c.WriteMessage(257, fmt.Sprintf("Created dir %s", p))
	} else {
		c.WriteMessage(550, fmt.Sprintf("Could not create %s : %v", p, err))
	}
}

func (c *Handler) handleRMD() {
	p := c.absPath(c.param)
	if err := c.driver.DeleteFile(c, p); err == nil {
		c.WriteMessage(250, fmt.Sprintf("Deleted dir %s", p))
	} else {
		c.WriteMessage(550, fmt.Sprintf("Could not delete dir %s: %v", p, err))
	}
}

func (c *Handler) handleCDUP() {
	parent, _ := path.Split(c.Path())
	if parent != "/" && strings.HasSuffix(parent, "/") {
		parent = parent[0 : len(parent)-1]
	}
	if err := c.driver.ChangeDirectory(c, parent); err == nil {
		c.SetPath(parent)
		c.WriteMessage(250, fmt.Sprintf("CDUP worked on %s", parent))
	} else {
		c.WriteMessage(550, fmt.Sprintf("CDUP issue: %v", err))
	}
}

func (c *Handler) handlePWD() {
	c.WriteMessage(257, "\""+c.Path()+"\" is the current directory")
}

func (c *Handler) handleLIST() {
	dir := c.absPath(c.param)

	if files, err := c.driver.ListFiles(c, dir); err == nil {
		if tr, err := c.TransferOpen(); err == nil {
			defer c.TransferClose()
			c.dirList(tr, files)
		}
	} else {
		c.WriteMessage(500, fmt.Sprintf("Could not list: %v", err))
	}
}

func fileStat(file os.FileInfo) string {
	return fmt.Sprintf(
		"%s 1 ftp ftp %12d %s %s",
		file.Mode(),
		file.Size(),
		file.ModTime().Format(" Jan _2 15:04 "),
		file.Name(),
	)
}

func (c *Handler) dirList(w io.Writer, files []os.FileInfo) error {
	for _, file := range files {
		fmt.Fprintf(w, "%s\r\n", fileStat(file))
	}
	fmt.Fprint(w, "\r\n")
	return nil
}
