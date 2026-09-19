package pantry

import (
	"fmt"

	"github.com/danielcbailey/Cookbook/core/models"
)

var volumeConversionsToMilliliters = map[models.IngredientUnit]float32{
	models.IngredientUnitMilliliters: 1,
	models.IngredientUnitLiters:      1000,
	models.IngredientUnitCups:        236.588,
	models.IngredientUnitTablespoons: 14.7868,
	models.IngredientUnitTeaspoons:   4.92892,
	models.IngredientUnitFluidOunces: 29.5735,
	models.IngredientUnitGallons:     3785.41,
	models.IngredientUnitQuarts:      946.353,
	models.IngredientUnitPints:       568.261,
}

var massConversionsToGrams = map[models.IngredientUnit]float32{
	models.IngredientUnitGrams:      1,
	models.IngredientUnitKilograms:  1000,
	models.IngredientUnitMilligrams: 0.001,
	models.IngredientUnitOunces:     28.3495,
	models.IngredientUnitPounds:     453.592,
}

func AddRecipeIngredients(acc models.RecipeIngredient, add models.RecipeIngredient) (models.RecipeIngredient, error) {
	qty, err := addIngredientQuantities(acc.Ingredient.Density, acc.Quantity, acc.Unit, add.Quantity, add.Unit)
	if err != nil {
		return models.RecipeIngredient{}, err
	}

	return models.RecipeIngredient{
		Quantity:   qty,
		Unit:       acc.Unit,
		Ingredient: acc.Ingredient,
	}, nil
}

func addIngredientQuantities(density float64, qtyAcc float32, unitAcc models.IngredientUnit, qtyAdd float32, unitAdd models.IngredientUnit) (float32, error) {
	addInDstUnit, err := ConvertIngredientUnit(density, qtyAdd, unitAdd, unitAcc)
	if err != nil {
		return 0, err
	}

	return addInDstUnit + qtyAcc, nil
}

func ConvertIngredientUnit(density float64, qtySrc float32, unitSrc, unitDst models.IngredientUnit) (float32, error) {
	srcVolConv, srcIsVol := volumeConversionsToMilliliters[unitSrc]
	dstVolConv, dstIsVol := volumeConversionsToMilliliters[unitDst]

	srcMassConv, srcIsMass := massConversionsToGrams[unitSrc]
	dstMassConv, dstIsMass := massConversionsToGrams[unitDst]

	switch {
	case srcIsVol && dstIsVol:
		ml := srcVolConv * qtySrc
		return ml / dstVolConv, nil
	case srcIsVol && dstIsMass:
		ml := srcVolConv * qtySrc
		g := float32(density) * ml
		return g / dstMassConv, nil
	case srcIsMass && dstIsVol:
		g := srcMassConv * qtySrc
		ml := g / float32(density)
		return ml / dstVolConv, nil
	case srcIsMass && dstIsMass:
		g := srcMassConv * qtySrc
		return g / dstMassConv, nil
	default:
		return 0, fmt.Errorf("cannot convert between %s and %s", unitSrc, unitDst)
	}
}
