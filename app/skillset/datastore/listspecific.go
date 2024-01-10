package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl SkillSetStorerImpl) ListByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*SkillSetPaginationListResult, error) {
	f := &SkillSetPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000,
		SortField: "sub_category",
		SortOrder: OrderAscending,
		TenantID:  tenantID,
		Status:    SkillSetStatusActive,
	}
	return impl.ListByFilter(ctx, f)
}
