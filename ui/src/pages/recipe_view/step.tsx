import { useState, type CSSProperties } from "react";
import type { RecipeIngredient, RecipeStep } from "../../apiTypes";
import { Chip, Typography } from "@heroui/react";
import { CentererdImage } from "../../shared/recipeCard";
import {Square, SquareCheck} from '@gravity-ui/icons';
import { quantityLabel } from "../../shared/recipeHelpers";
import { RecipeStepDescriptionText } from "./stepBody";

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
    flexDirection: 'row',
    gap: 10,
    paddingBottom: 10,
};

export function RecipeStepComponent({step}: {step: RecipeStep}) {
    const stepTextContainerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
        gap: 5,
    };

    return (
        <div style={stepStyle} className="shadow-surface">
            <div style={stepUpperContainerStyle}>
                {step.image_url !== '' && <CentererdImage src={step.image_url} alt={step.title} width={210} height={140} style={{borderRadius: 12}}/>}
                <div style={stepTextContainerStyle}>
                    <Typography type="h3" weight="semibold">{step.title}</Typography>
                    <RecipeStepDescriptionText step={step}/>
                </div>
            </div>
            
            <RecipeStepIngredients ingredients={step.ingredients}/>
        </div>
    );
}

function RecipeStepIngredients({ingredients}: {ingredients: RecipeIngredient[]}) {
    const containerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        gap: 20,
        flexWrap: "wrap",
    };

    return (
        <div style={containerStyle} className="no-select">
            {ingredients && ingredients.map((ingredient: RecipeIngredient) => {
                return <RecipeStepIngredientChip key={ingredient.index} ingr={ingredient}/>
            })}
        </div>
    );
}

export function RecipeStepIngredientChip({ingr}: {ingr: RecipeIngredient}) {
    const [used, setUsed] = useState<boolean>(false);

    return (
        <Chip size="lg" color="accent" variant={used ? 'soft' : 'secondary'} onClick={() => setUsed(!used)} style={{cursor: 'pointer'}}>
            {used ? <SquareCheck/> : <Square/>}
            {quantityLabel(ingr.quantity, ingr.unit) + ' ' + ingr.ingredient.name}
        </Chip>
    );
}

