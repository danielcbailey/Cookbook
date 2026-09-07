import { useEffect, useState, type CSSProperties } from "react";
import { Header } from "../../header/header";
import { Breadcrumbs, Button, ButtonGroup, IconPlus, Modal, Spinner, toast, Typography } from "@heroui/react";
import { getRecipe, RecipeAPIError, saveRecipe } from "../../recipeAPI";
import type { Ingredient, Recipe } from "../../apiTypes";
import { useNavigate, useParams } from "react-router-dom";
import { RecipeOverviewEdit } from "./overview";
import { RecipeEditStep } from "./step";
import { IngredientAPIError, listIngredients } from "../../ingredientAPI";
import { Camera, FloppyDisk, Sparkles } from "@gravity-ui/icons";
import { ImportModal } from "./importModal";


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

const toolsContainerStyle: CSSProperties = {
    width: 270,
    minWidth: 270,
};

export function RecipeEditPage() {
    const {id} = useParams<{id: string}>();
    const [recipe, setRecipe] = useState<Recipe | null>(!id ? blankRecipe() : null);
    const [ingredientOptions, setIngredientOptions] = useState<Ingredient[] | null>(null);
    const [pendingUpdate, setPendingUpdate] = useState<boolean>(false);
    const [importMode, setImportMode] = useState<'web' | 'photos' | null>(null);
    const [saving, setSaving] = useState<boolean>(false);

    const navigate = useNavigate();

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

    const addStep = () => {
        if (!recipe) return;

        const newRecipe: Recipe = {...recipe};
        if (!newRecipe.steps) {
            newRecipe.steps = [];
        }

        newRecipe.steps.push({
            id: 0,
            title: '',
            image_url: '',
            body_text: '',
        });

        setRecipe(newRecipe);
    }

    const onSaveRecipe = () => {
        if (!recipe) {
            return;
        }

        setSaving(true);
        saveRecipe(recipe).then((id: number) => {
            setSaving(false);
            navigate("/recipes/view/" + id.toString());
        }).catch((reason: RecipeAPIError) => {
            setSaving(false);
            toast.danger("Failed to Save Recipe", {
                description: reason.message,
            });
        });
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

                                    updateRecipe(newRecipe);
                                    return;
                                }

                                newRecipe.steps[idx] = newStep;
                                updateRecipe(newRecipe);
                            }}/>
                        })}
                        {!id && !pendingUpdate ? 
                            <GetStarted setImportMode={setImportMode} addStep={addStep}/> :
                            <Button onClick={addStep}>
                                <IconPlus style={{width: 16, height: 16}}/>
                                Add Step
                            </Button>}
                    </div>
                    <div style={toolsContainerStyle}>
                        <ButtonGroup variant="primary" fullWidth>
                            <Button isDisabled={!pendingUpdate} onClick={onSaveRecipe}>
                                <FloppyDisk style={{width: 16, height: 16}}/>
                                Save
                            </Button>
                        </ButtonGroup>
                    </div>
                </div>
            </div>
            <ImportModal onClose={() => setImportMode(null)} mode={importMode} onRecipeImport={(recipe) => {
                updateRecipe(recipe);
                setImportMode(null);
            }}/>
            <SavingModal isOpen={saving}/>
        </>
    );
}

function GetStarted({setImportMode, addStep}: {setImportMode: (mode: 'web' | 'photos') => void, addStep: () => void}) {
    return (<ButtonGroup variant="primary">
        <Button onClick={() => setImportMode('web')}>
            <Sparkles style={{width: 16, height: 16}}/>
            Web Import
        </Button>
        <Button onClick={() => setImportMode('photos')}>
            <ButtonGroup.Separator />
            <Camera style={{width: 16, height: 16}}/>
            Import From Photos
        </Button>
        <Button onClick={addStep}>
            <ButtonGroup.Separator />
            <IconPlus style={{width: 16, height: 16}}/>
            Add Step
        </Button>
    </ButtonGroup>);
}

function SavingModal({isOpen}: {isOpen: boolean}) {
    return (
        <Modal.Backdrop isOpen={isOpen}>
            <Modal.Container>
                <Modal.Dialog style={{display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center'}}>
                    <Typography type="h5" weight="semibold">
                        Saving Recipe
                    </Typography>
                    <Spinner size="lg"/>
                </Modal.Dialog>
            </Modal.Container>
        </Modal.Backdrop>
    )
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
            start_time: 0,
            unit: "second",
        },
        nutrition: {
            servings: 1,
            serving_mass: -1,
            serving_mass_unit: "gram",
            is_auto_generated: false,
            calories: -1,
            total_fat_grams: -1,
            saturated_fat_grams: -1,
            trans_fat_grams: -1,
            cholestrol_mg: -1,
            sodium_mg: -1,
            total_carbs_grams: -1,
            dietary_fiber_grams: -1,
            total_sugar_grams: -1,
            protein_grams: -1
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