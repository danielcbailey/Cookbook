import { useEffect, useState, type CSSProperties } from 'react';
import { Header } from '../../header/header';
import { RecipeSideBrowser } from './sideBrowser';
import type { RecipeListItem } from '../../apiTypes';
import { listCategories, listRecipesByCategory } from '../../recipeAPI';
import { toast } from '@heroui/react';
import { RecipeListRow } from './recipeRow';

const contentStyle: CSSProperties = {
    padding: 10,
    display: 'flex',
    flexDirection: 'row',
    width: '100%',
}

export function RecipesPage() {
    const [focusValue, setFocusValue] = useState<string | undefined>(undefined);
    const [recipes, setRecipes] = useState<Record<string,RecipeListItem[]>>({});

    useEffect(() => {
        const catchFn = (reason: unknown) => {
            if (reason instanceof Error) {
                toast.danger("Failed to Retrieve Recipes", {
                    description: reason.message,
                });
            }
        }

        listCategories().then((categories: string[]) => {
            for (const category of categories) {
                listRecipesByCategory(category, 0).then((respRecipes: RecipeListItem[]) => {
                    recipes['category/'+category] = respRecipes;
                    setRecipes({...recipes});
                }).catch(catchFn);
            }
        }).catch(catchFn);
    }, []);

    const rowsContainerStyle: CSSProperties = {
        width: '100%',
        // Without this the rows can push the parent flexbox wider than the viewport
        minWidth: 0,
        display: 'flex',
        flexDirection: 'column',
        gap: 10,
        paddingLeft: 20,
        paddingRight: 20,
        paddingTop: 5,
    };

    const rows = Object.keys(recipes);
    rows.sort(); // Ensures stable presentation order

    return (
        <>
            <Header/>
            <div style={contentStyle}>
                <RecipeSideBrowser focusValue={focusValue} focusRequest={setFocusValue} />
                <div style={rowsContainerStyle}>
                    {rows.map((rowKey: string) => {
                        return <RecipeListRow key={rowKey} recipes={recipes[rowKey]} label={rowKey}/>
                    })}
                </div>
            </div>
        </>
    );
}
