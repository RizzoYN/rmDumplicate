package ui

import (
	"os"
	"time"

	"github.com/ying32/govcl/vcl"
	"github.com/ying32/govcl/vcl/types"

	"rmDumplicate/fileHash"
)

var mapDumplicate = map[bool]string{
	true:  "是",
	false: "否",
}
var mapLoading = map[string]string{
	"分析":  "◑",
	"◑":  "◒",
	"◒":  "◐",
	"◐":  "◓",
	"◓":  "◑",
}

func NewCheckBox(parent vcl.IWinControl, title string, x, y, w, h int32) *vcl.TCheckBox {
	cb := vcl.NewCheckBox(parent)
	cb.SetParent(parent)
	cb.SetCaption(title)
	cb.SetBounds(x, y, w, h)
	return cb
}

type MainForm struct {
	*vcl.TForm
	pathEdit     *vcl.TEdit
	filesGrid    *vcl.TStringGrid
	button       *vcl.TButton
	sameDir      bool
	showHidden   bool
	mixMode      bool
	fileSelected []string
	defaultPath  string
}

func (m *MainForm) OnFormCreate(sender vcl.IObject) {
	m.SetCaption("重复文件查找")
	m.SetClientWidth(1080)
	m.SetClientHeight(640)
	m.EnabledMaximize(false)
	m.initComponents(m)
}

func (m *MainForm) initComponents(parent vcl.IWinControl) {
	popupMenu := vcl.NewPopupMenu(parent)
	cbOnTop := NewCheckBox(parent, "置顶", 20, 0, 20, 20)
	cbOnTop.SetOnClick(m.clickOnTop)
	cbDir := NewCheckBox(parent, "包含不同目录", 80, 0, 20, 20)
	cbDir.SetOnClick(m.clickSameDir)
	cbDir.SetHint("选中时不同路径下的文件如果具有相同的HASH值则计算为相同的文件")
	cbDir.SetOnMouseEnter(func(sender vcl.IObject) {
		vcl.AsButton(sender).ShowHint()
	})
	cbHid := NewCheckBox(parent, "显示隐藏文件", 190, 0, 20, 20)
	cbHid.SetOnClick(m.clickShowHidden)
	cbMix := NewCheckBox(parent, "混合模式", 295, 0, 20, 20)
	cbMix.SetOnClick(m.clickMixMode)
	cbMix.SetHint("计算文件HASH使用MD5算法，选中时计算文件HASH使用MD5与SHA256")
	cbMix.SetOnMouseEnter(func(sender vcl.IObject) {
		vcl.AsButton(sender).ShowHint()
	})
	m.sameDir = true
	m.showHidden = false
	m.showHidden = false
	m.pathEdit = vcl.NewEdit(parent)
	m.pathEdit.SetParent(m)
	m.pathEdit.SetBounds(20, 600, 800, 26)
	m.pathEdit.SetOnKeyUp(m.typedPathEdit)
	m.filesGrid = vcl.NewStringGrid(parent)
	m.filesGrid.SetParent(m)
	m.filesGrid.SetBounds(20, 25, 1040, 560)
	m.filesGrid.SetScrollBars(types.SsAutoVertical)
	m.filesGrid.AutoAdjustColumns()
	m.filesGrid.SetColCount(3)
	m.filesGrid.SetColWidths(0, 680)
	m.filesGrid.SetColWidths(1, 280)
	m.filesGrid.SetColWidths(2, 60)
	m.filesGrid.SetFixedCols(0)
	m.filesGrid.SetRowCount(1)
	m.filesGrid.SetPopupMenu(popupMenu)
	m.filesGrid.SetCells(0, 0, "文件路径")
	m.filesGrid.SetCells(1, 0, "文件MD5")
	m.filesGrid.SetCells(2, 0, "是否重复")
	m.button = vcl.NewButton(parent)
	m.button.SetParent(parent)
	m.button.SetCaption("选择目录")
	m.button.SetBounds(830, 600, 230, 26)
	m.button.SetOnClick(m.clickButton)
	m.defaultPath = `C\:`
}

func (m *MainForm) clickOnTop(sender vcl.IObject) {
	cb := vcl.AsCheckBox(sender)
	if cb.Checked() {
		m.SetFormStyle(types.FsSystemStayOnTop)
	} else {
		m.SetFormStyle(types.FsNormal)
	}
}

func (m *MainForm) clickSameDir(sender vcl.IObject) {
	cb := vcl.AsCheckBox(sender)
	if cb.Checked() {
		m.sameDir = false
	} else {
		m.sameDir = true
	}
}

func (m *MainForm) clickShowHidden(sender vcl.IObject) {
	cb := vcl.AsCheckBox(sender)
	if cb.Checked() {
		m.showHidden = false
	} else {
		m.showHidden = true
	}
}

func (m *MainForm) clickButton(sender vcl.IObject) {
	bt := vcl.AsButton(sender)
	var path string
	if cap := bt.Caption(); cap == "选择目录" || cap == "已完成 请选择目录" {
		ok, fileChoose := vcl.SelectDirectory2("选择目录", m.defaultPath, m.showHidden)
		if ok {
			bt.SetCaption("分析")
			m.pathEdit.SetTextBuf(fileChoose)
		} else {
			vcl.MessageDlg("无效的路径", types.MtError)
		}
	} else if bt.Caption() == "分析" {
		m.pathEdit.GetTextBuf(&path, 2147483647)
		bt.SetEnabled(false)
		go func() {
			fh := fileHash.NewFilesHash(path, m.sameDir, m.mixMode)
			if fh.ErrFlag {
				vcl.MessageDlg("文件异常", types.MtError)
				bt.SetCaption("选择目录")
				m.pathEdit.SetTextBuf("")
			} else {
				vcl.ThreadSync(func() {
					m.fileSelected = fh.FileSelected
					bt.SetCaption("删除重复文件")
					bt.SetEnabled(true)
					fileNums := len(fh.Files)
					m.filesGrid.SetRowCount(int32(fileNums) + 1)
					for i := 1; i <= fileNums; i++ {
						m.filesGrid.SetCells(0, int32(i), fh.Files[i-1])
						m.filesGrid.SetCells(1, int32(i), fh.FilesMD5[i-1])
						m.filesGrid.SetCells(2, int32(i), mapDumplicate[fh.FilesDumplicate[i-1]])
					}
				})
			}
		}()
		go func() {
			for {
				if bt.Enabled() {
					break
				} else {
					vcl.ThreadSync(func(){
						bt.SetCaption(mapLoading[bt.Caption()])
					 })
					time.Sleep(time.Millisecond * 80) 
				}
			}
		}()
	} else {
		if len(m.fileSelected) != 0 {
			for _, file := range m.fileSelected {
				os.Remove(file)
			}
		}
		m.button.SetCaption("已完成 请选择目录")
		m.pathEdit.SetTextBuf("")
		// clear有问题
		m.filesGrid.Clear()
	}
}

func (m *MainForm) typedPathEdit(sender vcl.IObject, key *types.Char, shift types.TShiftState) {
	pathEdit := vcl.AsEdit(sender)
	pathLength := pathEdit.GetTextLen()
	if pathLength == 0 {
		m.button.SetCaption("选择目录")
	} else {
		m.button.SetCaption("分析")
	}
}

func (m *MainForm) clickMixMode(sender vcl.IObject) {
	cb := vcl.AsCheckBox(sender)
	if cb.Checked() {
		m.mixMode = true
	} else {
		m.mixMode = false
	}
}
