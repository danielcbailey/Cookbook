import { type CSSProperties } from "react";
import type { Ingredient, RecipeIngredient, RecipeStep } from "../../apiTypes";
import { CentererdImage } from "../../shared/recipeCard";
import { Button, Chip, IconPlus, Input, TextField } from "@heroui/react";
import { getStepIngredients, ingredientStepString, quantityLabel } from "../../shared/recipeHelpers";
import { Clock, TrashBin } from "@gravity-ui/icons";
import { IngredientPopup } from "./ingredient_popup";
import { ImageUpload } from "../../shared/imageUpload";
import { RecipeStepEditDescriptionText } from "./stepBody";


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

export function RecipeEditStep({step, setStep, ingredientOptions}: {step: RecipeStep, setStep: (r: RecipeStep | null) => void, ingredientOptions: Ingredient[]}) {
    const stepTextContainerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
        gap: 5,
        width: '100%',
    };

    const addIngredient = (ingr: RecipeIngredient) => {
        const newStep = {...step};
        newStep.body_text += ingredientStepString(ingr);
        setStep(newStep);
    }

    const onImageUpload = (dataURI: string) => {
        const newStep: RecipeStep = {...step};
        newStep.image_url = dataURI;
        setStep(newStep);
    }

    return (
        <div style={stepStyle} className="shadow-surface">
            <div style={stepUpperContainerStyle}>
                <ImageUpload width={210} height={140} onImage={onImageUpload} maxArea={1000 * 1000}>
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
                    
                    <RecipeStepEditDescriptionText step={step} setStep={setStep}/>
                </div>
            </div>
            
            <RecipeStepEditIngredients ingredients={getStepIngredients(step)} options={ingredientOptions} addIngredient={addIngredient}/>
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
                return <RecipeEditStepIngredientChip key={ingredientStepString(ingredient)} ingr={ingredient}/>
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