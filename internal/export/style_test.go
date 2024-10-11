package export

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStyle(t *testing.T) {
	assert := assert.New(t)

	styles := newStyles(
		newStyleItem(1, 1),
		newStyleItem("2", 2),
		newStyleItem('3', 3),
		newStyleItem(Positive, 4),
		newStyleItem(Negative, 5),
		newStyleItem(Header, 6),
	)

	assert.Equal(1, styles.style(1))
	assert.Equal(2, styles.style("2"))
	assert.Equal(3, styles.style('3'))
	assert.Equal(4, styles.style("true"))
	assert.Equal(5, styles.style("false"))
	assert.Equal(4, styles.style("t"))
	assert.Equal(5, styles.style("f"))
	assert.Equal(4, styles.style("T"))
	assert.Equal(5, styles.style("F"))
	assert.Equal(4, styles.style("是"))
	assert.Equal(5, styles.style("否"))
	assert.Equal(4, styles.style("True"))
	assert.Equal(5, styles.style("False"))
	assert.Equal(4, styles.style("TRUE"))
	assert.Equal(5, styles.style("FALSE"))
	assert.Equal(6, styles.style(Header))
	assert.Equal(0, styles.style("abc"))
	assert.Equal(0, styles.style(90))
}
