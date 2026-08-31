import { useEffect, useState, type CSSProperties } from "react";
import { Header } from "../../header/header";
import { Breadcrumbs, toast } from "@heroui/react";
import { getRecipe, RecipeAPIError } from "../../recipeAPI";
import type { Ingredient, Recipe } from "../../apiTypes";
import { useParams } from "react-router-dom";
import { RecipeOverviewEdit } from "./overview";
import { RecipeEditStep } from "./step";
import { IngredientAPIError, listIngredients } from "../../ingredientAPI";


const contentStyle: CSSProperties = {
    padding: 10,
    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 30,
};

const pageStyle: CSSProperties = {
    paddingLeft: 30,
    paddingRight: 30,
    paddingTop: 5,
    width: '100%',
    height: '100%',
    maxHeight: '100%',
};

const stepsContainerStyle: CSSProperties = {
    width: '100%',
    display: 'flex',
    flexDirection: 'column',
    maxHeight: '100%',
    gap: 10,
};

export function RecipeEditPage() {
    const {id} = useParams<{id: string}>();
    const [recipe, setRecipe] = useState<Recipe | null>(!id ? blankRecipe() : null);
    const [ingredientOptions, setIngredientOptions] = useState<Ingredient[] | null>(null);
    const [pendingUpdate, setPendingUpdate] = useState<boolean>(false);

    useEffect(() => {
        if (!id) {
            return;
        }

        getRecipe(parseInt(id || '')).then((recipe: Recipe) => {
            setRecipe(recipe);
        }).catch((reason: RecipeAPIError) => {
            toast.danger("Failed to Retrieve Recipe", {
                description: reason.message,
            });
        });

        listIngredients(0, 0).then((ingredients: Ingredient[]) => {
            setIngredientOptions(ingredients);
        }).catch((reason: IngredientAPIError) => {
            toast.danger("Failed to Retrieve Ingredients", {
                description: reason.message,
            });
        });
    }, [id]);

    const updateRecipe = (r: Recipe) => {
        setRecipe(r);
        setPendingUpdate(true);
    }

    return (
        <>
            <Header/>
            <div style={pageStyle}>
                <Breadcrumbs>
                    <Breadcrumbs.Item href="/recipes">Recipes</Breadcrumbs.Item>
                    <Breadcrumbs.Item>{!id ? "Create Recipe" : "Edit Recipe"}</Breadcrumbs.Item>
                </Breadcrumbs>
                <div style={contentStyle}>
                    {recipe && <RecipeOverviewEdit recipe={recipe} setRecipe={updateRecipe}/>}
                    <div style={stepsContainerStyle}>
                        {recipe && recipe.steps.map((step, idx) => {
                            return <RecipeEditStep key={idx} step={step} ingredientOptions={ingredientOptions ?? []} setStep={(newStep) => {
                                const newRecipe: Recipe = {...recipe};

                                if (newStep == null) {
                                    // Remove the step
                                    newRecipe.steps.splice(idx, 1);
                                    for (let i = idx; i < newRecipe.steps.length; i++) {
                                        newRecipe.steps[i].index = i;
                                    }
                                    updateRecipe(newRecipe);
                                    return;
                                }

                                newRecipe.steps[idx] = newStep;
                                updateRecipe(newRecipe);
                            }}/>
                        })}
                    </div>
                    {pendingUpdate.toString()}
                </div>
            </div>
        </>
    );
}

function blankRecipe(): Recipe {
    return {
        id: 0,
        user_id: 0,
        title: "",
        description: "",
        image_url: "",
        author: "",
        publisher: "",
        servings: 0,
        ingredients: [],
        steps: [],
        total_time: {
            index: 0,
            start_time: 0,
            unit: "second",
        },
        nutrition: {
            servings: 1,
            serving_mass: 0,
            serving_mass_unit: "gram",
            is_auto_generated: false,
            calories: 0,
            total_fat_grams: 0,
            saturated_fat_grams: 0,
            trans_fat_grams: 0,
            cholestrol_mg: 0,
            sodium_mg: 0,
            total_carbs_grams: 0,
            dietary_fiber_grams: 0,
            total_sugar_grams: 0,
            protein_grams: 0
        },
        category: "",
        protein: "",
        cuisine: "",
        suggested_meal: "dinner",
        tags: [],
        created_at: 0,
        updated_at: 0,
    };
}