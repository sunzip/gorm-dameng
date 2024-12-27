package dameng

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DmIndexHint struct {
	Type string
	Keys []string
}

// 实现 GORM 的 Clause 接口
func (h DmIndexHint) Name() string {
	return "INDEX_HINT"
}

func (h DmIndexHint) Build(builder clause.Builder) {
	if len(h.Keys) > 0 {
		builder.WriteString("/*+ ")
		switch strings.Trim(h.Type, " ") {
		case "USE INDEX":
			builder.WriteString("INDEX")
		case "IGNORE INDEX":
			builder.WriteString("NO_INDEX")
		case "FORCE INDEX":
			builder.WriteString("FORCE_INDEX")
		default:
			builder.WriteString("INDEX")
			return
		}
		builder.WriteByte('(')
		if s, ok := builder.(*gorm.Statement); ok {
			builder.WriteString(s.Table)
		}
		builder.WriteByte(' ')
		for i, index := range h.Keys {
			if i > 0 {
				builder.WriteByte(',')
			}
			builder.WriteQuoted(index)
		}
		builder.WriteString(") */ ")
	}
}

func IndexHintFromClauseBuilder(c clause.Clause, builder clause.Builder) {
	if c.BeforeExpression != nil {
		c.BeforeExpression.Build(builder)
		builder.WriteByte(' ')
	}

	if c.Name != "" {
		builder.WriteString(c.Name)
		builder.WriteByte(' ')
	}

	if c.AfterNameExpression != nil {
		c.AfterNameExpression.Build(builder)
		builder.WriteByte(' ')
	}

	if from, ok := c.Expression.(clause.From); ok {
		joins := from.Joins
		from.Joins = nil
		from.Build(builder)

		// set indexHints in the middle between table and joins
		squashExpression(c.AfterExpression, func(expression clause.Expression) {
			if indexHint, ok := expression.(DmIndexHint); ok { // pick
				builder.WriteByte(' ')
				indexHint.Build(builder)
			}
		})

		for _, join := range joins {
			builder.WriteByte(' ')
			join.Build(builder)
		}
	} else {
		c.Expression.Build(builder)
	}

	squashExpression(c.AfterExpression, func(expression clause.Expression) {
		if _, ok := expression.(DmIndexHint); ok {
			return
		}
		builder.WriteByte(' ')
		expression.Build(builder)
	})
}

// ///
type Exprs []clause.Expression

func (exprs Exprs) Build(builder clause.Builder) {
	for idx, expr := range exprs {
		if idx > 0 {
			builder.WriteByte(' ')
		}
		expr.Build(builder)
	}
}

func squashExpression(expression clause.Expression, do func(expression clause.Expression)) {
	if exprs, ok := expression.(Exprs); ok {
		for _, expr := range exprs {
			squashExpression(expr, do)
		}
	} else if expression != nil {
		do(expression)
	}
}
