package listviewport

import (
	"time"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type Model struct {
	ctx             context.ProgramContext
	viewport        viewport.Model
	topBoundId      int
	bottomBoundId   int
	currId          int
	ListItemHeight  int
	NumCurrentItems int
	NumTotalItems   int
	LastUpdated     time.Time
	CreatedAt       time.Time
	ItemTypeLabel   string
}

func NewModel(
	ctx context.ProgramContext,
	dimensions constants.Dimensions,
	lastUpdated time.Time,
	createdAt time.Time,
	itemTypeLabel string,
	numItems, listItemHeight int,
) Model {
	model := Model{
		ctx:             ctx,
		NumCurrentItems: numItems,
		ListItemHeight:  listItemHeight,
		currId:          0,
		viewport: viewport.New(
			viewport.WithWidth(dimensions.Width),
			viewport.WithHeight(dimensions.Height),
		),
		topBoundId:    0,
		ItemTypeLabel: itemTypeLabel,
		LastUpdated:   lastUpdated,
		CreatedAt:     createdAt,
	}
	model.bottomBoundId = utils.Min(
		model.NumCurrentItems-1,
		model.getNumPrsPerPage()-1,
	)
	return model
}

func (m *Model) SetNumItems(numItems int) {
	m.NumCurrentItems = numItems
	m.bottomBoundId = utils.Min(m.NumCurrentItems-1, m.getNumPrsPerPage()-1)
}

func (m *Model) SetTotalItems(total int) {
	m.NumTotalItems = total
}

func (m *Model) SetItemHeight(height int) {
	m.ListItemHeight = height
}

func (m *Model) SyncViewPort(content string) {
	m.viewport.SetContent(content)
}

func (m *Model) getNumPrsPerPage() int {
	if m.ListItemHeight == 0 {
		return 0
	}
	return m.viewport.Height() / m.ListItemHeight
}

func (m *Model) ResetCurrItem() {
	m.currId = 0
	m.viewport.GotoTop()
}

func (m *Model) GetCurrItem() int {
	return m.currId
}

func (m *Model) NextItem() int {
	m.snapToCursor()

	atBottomOfViewport := m.currId >= m.bottomBoundId
	if atBottomOfViewport {
		m.topBoundId += 1
		m.bottomBoundId += 1
		m.viewport.ScrollDown(m.ListItemHeight)
	}

	newId := utils.Min(m.currId+1, m.NumCurrentItems-1)
	newId = utils.Max(newId, 0)
	m.currId = newId
	return m.currId
}

func (m *Model) PrevItem() int {
	m.snapToCursor()

	if m.currId > 0 && m.currId <= m.topBoundId {
		m.topBoundId -= 1
		m.bottomBoundId -= 1
		m.viewport.ScrollUp(m.ListItemHeight)
	}

	m.currId = utils.Max(m.currId-1, 0)
	return m.currId
}

func (m *Model) SetCurrItem(id int) {
	if m.NumCurrentItems == 0 {
		return
	}
	m.currId = utils.Clamp(0, id, m.NumCurrentItems-1)
	m.snapToCursor()
}

// snapToCursor scrolls the viewport so that the currently selected item is
// visible. If it's already visible, this is a no-op.
func (m *Model) snapToCursor() {
	if m.ListItemHeight == 0 {
		return
	}
	perPage := m.getNumPrsPerPage()
	if perPage == 0 {
		return
	}
	// Bounds, not viewport.YOffset(), are the source of truth here: the wheel
	// handlers sync bounds from the viewport, but a viewport with no content
	// clamps YOffset to 0 and would report every item as visible.
	if m.currId < m.topBoundId {
		m.topBoundId = m.currId
	} else if m.currId > m.bottomBoundId {
		m.topBoundId = m.currId - perPage + 1
	} else {
		return
	}
	m.bottomBoundId = utils.Min(m.NumCurrentItems-1, m.topBoundId+perPage-1)
	m.viewport.SetYOffset(m.topBoundId * m.ListItemHeight)
}

func (m *Model) ScrollUp(items int) {
	if m.ListItemHeight == 0 || items <= 0 {
		return
	}
	m.viewport.ScrollUp(items * m.ListItemHeight)
	m.syncBoundsFromViewport()
}

func (m *Model) ScrollDown(items int) {
	if m.ListItemHeight == 0 || items <= 0 {
		return
	}
	m.viewport.ScrollDown(items * m.ListItemHeight)
	m.syncBoundsFromViewport()
}

func (m *Model) syncBoundsFromViewport() {
	if m.ListItemHeight == 0 {
		return
	}
	m.topBoundId = m.viewport.YOffset() / m.ListItemHeight
	perPage := m.getNumPrsPerPage()
	m.bottomBoundId = utils.Min(m.NumCurrentItems-1, m.topBoundId+perPage-1)
}

func (m *Model) FirstItem() int {
	m.currId = 0
	m.viewport.GotoTop()
	return m.currId
}

func (m *Model) LastItem() int {
	m.currId = m.NumCurrentItems - 1
	m.viewport.GotoBottom()
	return m.currId
}

func (m *Model) SetDimensions(dimensions constants.Dimensions) {
	m.viewport.SetHeight(max(0, dimensions.Height))
	m.viewport.SetWidth(max(0, dimensions.Width))
}

func (m *Model) View() string {
	viewport := m.viewport.View()
	return lipgloss.NewStyle().
		Width(m.viewport.Width()).
		MaxWidth(m.viewport.Width()).
		Render(
			viewport,
		)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = *ctx
}
