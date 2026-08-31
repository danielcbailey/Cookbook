import { type CSSProperties } from "react";
import type { Ingredient, RecipeIngredient, RecipeStep } from "../../apiTypes";
import { CentererdImage } from "../../shared/recipeCard";
import { ImageUpload } from "./overview";
import { Button, Chip, IconPlus, Input, TextField } from "@heroui/react";
import { RecipeStepDescriptionText } from "../recipe_view/stepBody";
import { quantityLabel } from "../../shared/recipeHelpers";
import { Clock, TrashBin } from "@gravity-ui/icons";
import { IngredientPopup } from "./ingredient_popup";


const stepStyle: CSSProperties = {
    width: '100%',
    display: 'flex',
    flexDirection: 'column',
    gap: 10,
    padding: 10,
    borderRadius: 16,
    backgroundColor: 'var(--surface)',
};

const stepUpperContainerStyle: CSSProperties = {
    display: 'flex',
    width: '100%',
    flexDirection: 'row',
    gap: 10,
    paddingBottom: 10,
};

const stepHeaderStyle: CSSProperties = {
    display: 'flex',
    flexDirection: 'row',
    gap: 10,
    alignItems: 'flex-start',
};

function reassignIngredientIndices(step: RecipeStep) {
    const bodyIngredients: RecipeIngredient[] = [];
    const INGR_PLACEHOLDER = /\{\{ingr-(\d+)\}\}/g;
    const indices = [...step.body_text.matchAll(INGR_PLACEHOLDER)].map((m) => Number(m[1]));
    const indexSet: Record<number, boolean> = {};

    for (const idx of indices) {
        const ingr = step.ingredients[idx];
        if (indexSet[idx]) {
            ingr.id = 0;
        }

        ingr.index = bodyIngredients.length;
        bodyIngredients.push(ingr);
    }

    step.ingredients = bodyIngredients;
}

export function RecipeEditStep({step, setStep, ingredientOptions}: {step: RecipeStep, setStep: (r: RecipeStep | null) => void, ingredientOptions: Ingredient[]}) {
    const stepTextContainerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
        gap: 5,
        width: '100%',
    };

    const stepTextStyle: CSSProperties = {
        width: '100%',
        minHeight: 100,
        backgroundColor: 'var(--default)',
        borderRadius: 12,
        paddingTop: 8,
        paddingBottom: 8,
        paddingLeft: 12,
        paddingRight: 12,
    };

    const addIngredient = (ingr: RecipeIngredient) => {
        const newStep: RecipeStep = {...step};
        ingr.index = newStep.ingredients.length;
        newStep.ingredients.push(ingr);

        newStep.body_text += "{{ingr-" + ingr.index.toString() + "}}";
        reassignIngredientIndices(newStep);
        setStep(newStep);
    }

    return (
        <div style={stepStyle} className="shadow-surface">
            <div style={stepUpperContainerStyle}>
                <ImageUpload width={210} height={140}>
                    {step.image_url !== '' && <CentererdImage src={step.image_url} alt={step.title} width={210} height={140} style={{borderRadius: 12}}/>}
                </ImageUpload>
                
                <div style={stepTextContainerStyle}>
                    <div style={stepHeaderStyle}>
                        <TextField fullWidth variant="secondary">
                            <Input value={step.title} placeholder="Step Title" onChange={(e) => {
                                const newStep = {...step};
                                newStep.title = e.currentTarget.value;
                                setStep(newStep);
                            }}/>
                        </TextField>
                        <Button isIconOnly variant="secondary" style={{minWidth: 36}}>
                            <Clock/>
                        </Button>
                        <Button isIconOnly variant="danger" onClick={() => {setStep(null)}} style={{minWidth: 36}}>
                            <TrashBin />
                        </Button>
                    </div>
                    
                    <RecipeStepDescriptionText step={step} style={stepTextStyle} onEdit={(v) => {
                        const newStep: RecipeStep = {...step};
                        newStep.body_text = v;
                        reassignIngredientIndices(newStep);
                        setStep(newStep);
                    }}/>
                </div>
            </div>
            
            <RecipeStepEditIngredients ingredients={step.ingredients} options={ingredientOptions} addIngredient={addIngredient}/>
        </div>
    );
}

function RecipeStepEditIngredients({ingredients, options, addIngredient}: {ingredients: RecipeIngredient[], options: Ingredient[], addIngredient: (ingr: RecipeIngredient) => void}) {
    const containerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        gap: 20,
        flexWrap: "wrap",
    };

    return (
        <div style={containerStyle} className="no-select">
            {ingredients && ingredients.map((ingredient: RecipeIngredient) => {
                return <RecipeEditStepIngredientChip key={ingredient.index} ingr={ingredient}/>
            })}
            <IngredientPopup options={options} allowCreate onSelect={(r: RecipeIngredient) => addIngredient(r)}>
                <RecipeEditStepIngredientChip/>
            </IngredientPopup>
        </div>
    );
}

function RecipeEditStepIngredientChip({ingr, onClick}: {ingr?: RecipeIngredient, onClick?: () => void}) {
    return (
        <Chip size="lg" color="accent" variant={'soft'} onClick={() => onClick?.()} style={{cursor: 'pointer'}}>
            {!ingr && <IconPlus/>}
            {ingr && (quantityLabel(ingr.quantity, ingr.unit) + ' ' + ingr.ingredient.name)}
        </Chip>
    );
}