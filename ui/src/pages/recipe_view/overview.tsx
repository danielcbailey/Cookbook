import { type CSSProperties } from 'react'
import { CentererdImage, RecipeTag } from '../../shared/recipeCard'
import { timeUnitLabel } from '../../shared/recipeHelpers'
import { capitalizeWords } from '../../helpers'
import type { Recipe } from '../../apiTypes'
import { Typography } from '@heroui/react'

const overviewWidth = 450;

export function RecipeOverview({recipe}: {recipe: Recipe}) {
    const overviewStyle: CSSProperties = {
        width: overviewWidth,
        minWidth: overviewWidth,
        flexGrow: 0,
        display: 'flex',
        flexDirection: 'column',
        gap: 10,
    };

    const statGroupStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        justifyContent: 'space-between',
    };

    const statGroupDividerStyle: CSSProperties = {
        width: 1,
        minWidth: 1,
        backgroundColor: 'var(--border)',
        height: '100%',
    }

    const prepTimeString = recipe.total_time.start_time.toString() +
        (recipe.total_time.end_time ? ' - ' + recipe.total_time.end_time.toString() : '') +
        ' ' + timeUnitLabel(recipe.total_time.unit);

    const tagContainerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        gap: 5,
    };

    return (
        <div style={overviewStyle}>
            <CentererdImage
                src={recipe.image_url}
                alt={recipe.title}
                width={overviewWidth}
                height={overviewWidth * 2/3}
                style={{borderRadius: 16}}/>

            <Typography type="h2" weight="bold">{recipe.title}</Typography>

            <div style={statGroupStyle}>
                <RecipeStat label="Prep Time" value={prepTimeString}/>
                <div style={statGroupDividerStyle}/>
                <RecipeStat label="Calories" value={recipe.nutrition.calories.toString()}/>
                <div style={statGroupDividerStyle}/>
                <RecipeStat label="Servings" value={recipe.servings.toString()}/>
            </div>

            <Typography type="body-sm" style={{lineHeight: '20px'}}>{recipe.description}</Typography>

            <div style={tagContainerStyle}>
                {recipe.tags && recipe.tags.map((tag) => {
                    return <RecipeTag tag={tag.name}/>
                })}
            </div>

            <div style={statGroupStyle}>
                <RecipeStat label="Category" value={capitalizeWords(recipe.category)}/>
                <div style={statGroupDividerStyle}/>
                <RecipeStat label="Protein" value={capitalizeWords(recipe.protein)}/>
                <div style={statGroupDividerStyle}/>
                <RecipeStat label="Suggested Meal" value={capitalizeWords(recipe.suggested_meal)}/>
            </div>
        </div>
    );
}

function RecipeStat({label, value}: {label: string, value: string}) {
    const statStyle: CSSProperties = {
        width: '100%',
        display: 'flex',
        flexDirection: 'column',
        gap: 0,
        alignItems: 'center',
    };

    return (
        <div style={statStyle}>
            <Typography color="muted">{label}</Typography>
            <Typography weight="semibold">{value}</Typography>
        </div>
    );
}