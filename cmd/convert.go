package cmd

import (
	a_s "github.com/over55/workery-cli/app/associate/datastore"
	c_s "github.com/over55/workery-cli/app/customer/datastore"
	o_s "github.com/over55/workery-cli/app/order/datastore"
	ti_s "github.com/over55/workery-cli/app/taskitem/datastore"
)

// ------------------------------------------------ ORDER ------------------------------------------------ //

// By Customer

func toOrderTagsFromCustomerTags(tt []*c_s.CustomerTag) []*o_s.OrderTag {
	var titt []*o_s.OrderTag = make([]*o_s.OrderTag, 0)
	for _, t := range tt {
		titt = append(titt, &o_s.OrderTag{
			ID:          t.ID,
			Text:        t.Text,
			Description: t.Description,
		})
	}
	return titt
}

func toOrderSkillSet(sss []*o_s.OrderSkillSet) []*o_s.OrderSkillSet {
	var tiss []*o_s.OrderSkillSet = make([]*o_s.OrderSkillSet, 0)
	for _, ss := range sss {
		tiss = append(tiss, &o_s.OrderSkillSet{
			ID:          ss.ID,
			Category:    ss.Category,
			SubCategory: ss.SubCategory,
			Description: ss.Description,
			Status:      ss.Status,
		})
	}
	return tiss
}

func toOrderInsuranceRequirementsFromCustomerTags(tt []*a_s.AssociateInsuranceRequirement) []*o_s.OrderInsuranceRequirement {
	var arr []*o_s.OrderInsuranceRequirement = make([]*o_s.OrderInsuranceRequirement, 0)
	for _, t := range tt {
		arr = append(arr, &o_s.OrderInsuranceRequirement{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Status:      t.Status,
		})
	}
	return arr
}

func toOrderSkillSetsFromCustomerTags(tt []*a_s.AssociateSkillSet) []*o_s.OrderSkillSet {
	var arr []*o_s.OrderSkillSet = make([]*o_s.OrderSkillSet, 0)
	for _, t := range tt {
		arr = append(arr, &o_s.OrderSkillSet{
			ID:          t.ID,
			Category:    t.Category,
			SubCategory: t.SubCategory,
			Description: t.Description,
			Status:      t.Status,
		})
	}
	return arr
}

// By Associate

func toOrderTagsFromAssociateTags(fromArr []*a_s.AssociateTag) []*o_s.OrderTag {
	var toArr []*o_s.OrderTag = make([]*o_s.OrderTag, 0)
	for _, t := range fromArr {
		toArr = append(toArr, &o_s.OrderTag{
			ID:          t.ID,
			Text:        t.Text,
			Description: t.Description,
			Status:      t.Status,
		})
	}
	return toArr
}

func toOrderSkillSetsFromAssociateSkillSets(fromArr []*a_s.AssociateSkillSet) []*o_s.OrderSkillSet {
	var toArr []*o_s.OrderSkillSet = make([]*o_s.OrderSkillSet, 0)
	for _, ss := range fromArr {
		toArr = append(toArr, &o_s.OrderSkillSet{
			ID:          ss.ID,
			Category:    ss.Category,
			SubCategory: ss.SubCategory,
			Description: ss.Description,
			Status:      ss.Status,
		})
	}
	return toArr
}

func toOrderInsuranceRequirementsFromAssociateInsuranceRequirements(fromArr []*a_s.AssociateInsuranceRequirement) []*o_s.OrderInsuranceRequirement {
	var toArr []*o_s.OrderInsuranceRequirement = make([]*o_s.OrderInsuranceRequirement, 0)
	for _, t := range fromArr {
		toArr = append(toArr, &o_s.OrderInsuranceRequirement{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Status:      t.Status,
		})
	}
	return toArr
}

func toOrderVehicleTypesFromAssociateVehicleTypes(fromArr []*a_s.AssociateVehicleType) []*o_s.OrderVehicleType {
	var toArr []*o_s.OrderVehicleType = make([]*o_s.OrderVehicleType, 0)
	for _, t := range fromArr {
		toArr = append(toArr, &o_s.OrderVehicleType{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Status:      t.Status,
		})
	}
	return toArr
}

// ------------------------------------------------ TASK ITEM ------------------------------------------------ //

// By Customer

func toTaskItemTagsFromCustomerTags(tt []*c_s.CustomerTag) []*ti_s.TaskItemTag {
	var titt []*ti_s.TaskItemTag = make([]*ti_s.TaskItemTag, 0)
	for _, t := range tt {
		titt = append(titt, &ti_s.TaskItemTag{
			ID:          t.ID,
			Text:        t.Text,
			Description: t.Description,
			Status:      t.Status,
		})
	}
	return titt
}

func toTaskItemSkillSetsFromOrderSkillSets(sss []*o_s.OrderSkillSet) []*ti_s.TaskItemSkillSet {
	var tiss []*ti_s.TaskItemSkillSet = make([]*ti_s.TaskItemSkillSet, 0)
	for _, ss := range sss {
		tiss = append(tiss, &ti_s.TaskItemSkillSet{
			ID:          ss.ID,
			Category:    ss.Category,
			SubCategory: ss.SubCategory,
			Description: ss.Description,
			Status:      ss.Status,
		})
	}
	return tiss
}

func toTaskItemTagsFromOrderTags(tt []*o_s.OrderTag) []*ti_s.TaskItemTag {
	var titt []*ti_s.TaskItemTag = make([]*ti_s.TaskItemTag, 0)
	for _, t := range tt {
		titt = append(titt, &ti_s.TaskItemTag{
			ID:          t.ID,
			Text:        t.Text,
			Description: t.Description,
			Status:      t.Status,
		})
	}
	return titt
}

// By Associate

func toTaskItemTagsFromAssociateTags(fromArr []*a_s.AssociateTag) []*ti_s.TaskItemTag {
	var toArr []*ti_s.TaskItemTag = make([]*ti_s.TaskItemTag, 0)
	for _, t := range fromArr {
		toArr = append(toArr, &ti_s.TaskItemTag{
			ID:          t.ID,
			Text:        t.Text,
			Description: t.Description,
			Status:      t.Status,
		})
	}
	return toArr
}

func toTaskItemSkillSetsFromAssociateSkillSets(fromArr []*a_s.AssociateSkillSet) []*ti_s.TaskItemSkillSet {
	var toArr []*ti_s.TaskItemSkillSet = make([]*ti_s.TaskItemSkillSet, 0)
	for _, ss := range fromArr {
		toArr = append(toArr, &ti_s.TaskItemSkillSet{
			ID:          ss.ID,
			Category:    ss.Category,
			SubCategory: ss.SubCategory,
			Description: ss.Description,
			Status:      ss.Status,
		})
	}
	return toArr
}

func toTaskItemInsuranceRequirementsFromAssociateInsuranceRequirements(fromArr []*a_s.AssociateInsuranceRequirement) []*ti_s.TaskItemInsuranceRequirement {
	var toArr []*ti_s.TaskItemInsuranceRequirement = make([]*ti_s.TaskItemInsuranceRequirement, 0)
	for _, t := range fromArr {
		toArr = append(toArr, &ti_s.TaskItemInsuranceRequirement{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Status:      t.Status,
		})
	}
	return toArr
}

func toTaskItemVehicleTypesFromAssociateVehicleTypes(fromArr []*a_s.AssociateVehicleType) []*ti_s.TaskItemVehicleType {
	var toArr []*ti_s.TaskItemVehicleType = make([]*ti_s.TaskItemVehicleType, 0)
	for _, t := range fromArr {
		toArr = append(toArr, &ti_s.TaskItemVehicleType{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Status:      t.Status,
		})
	}
	return toArr
}
