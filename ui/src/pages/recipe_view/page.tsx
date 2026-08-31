import { useEffect, useState, type CSSProperties } from 'react'
import { Breadcrumbs, Button, ButtonGroup, toast } from '@heroui/react'
import { Header } from '../../header/header'
import { useNavigate, useParams } from 'react-router-dom'
import type { Recipe } from '../../apiTypes'
import { getRecipe, RecipeAPIError } from '../../recipeAPI'
import { RecipeOverview } from './overview'
import { RecipeStepComponent } from './step'
import {PencilToSquare, TrashBin} from '@gravity-ui/icons';

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
    border: '1px solid green',
};

export function RecipeViewPage({parent}: {parent: string}) {
    const {id} = useParams<{id: string}>();
    const [recipe, setRecipe] = useState<Recipe | null>(null);
    const navigate = useNavigate();

    useEffect(() => {
        getRecipe(parseInt(id || '')).then((recipe: Recipe) => {
            setRecipe(recipe);
        }).catch((reason: RecipeAPIError) => {
            toast.danger("Failed to Retrieve Recipe", {
                description: reason.message,
            });
        });
    }, [id]);

    return (
        <>
            <Header/>
            <div style={pageStyle}>
                <Breadcrumbs>
                    <Breadcrumbs.Item href={"/"+parent.toLocaleLowerCase()}>{parent}</Breadcrumbs.Item>
                    <Breadcrumbs.Item>View Recipe</Breadcrumbs.Item>
                </Breadcrumbs>
                <div style={contentStyle}>
                    {recipe && <RecipeOverview recipe={recipe}/>}

                    <div style={stepsContainerStyle}>
                        {recipe && recipe.steps.map((step) => {
                            return <RecipeStepComponent step={step}/>
                        })}
                    </div>

                    <div style={toolsContainerStyle}>
                        <ButtonGroup variant="primary" fullWidth>
                            <Button onClick={() => navigate("/recipes/edit/" + id?.toString())}>
                                <PencilToSquare style={{width: 16, height: 16}}/>
                                Edit
                            </Button>
                            <Button>
                                <ButtonGroup.Separator />
                                <TrashBin style={{width: 16, height: 16}}/>
                                Delete
                            </Button>
                        </ButtonGroup>
                    </div>
                </div>
            </div>
        </>
    );
}