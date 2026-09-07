import { useState, type CSSProperties } from "react";
import type { Recipe } from "../../apiTypes";
import { CentererdImage, RecipeTag } from "../../shared/recipeCard";
import { Description, Input, Label, TextArea, TextField, toast } from "@heroui/react";
import { PopupTypeSelect, type PopupTypeSelectOption } from "../../shared/popupTypeSelect";
import { listTags, RecipeAPIError } from "../../recipeAPI";
import { ImageUpload } from "../../shared/imageUpload";


const overviewWidth = 450;
const descriptionLimitChars = 400;

export function RecipeOverviewEdit({recipe, setRecipe}: {recipe: Recipe, setRecipe: (r: Recipe) => void}) {
    const overviewStyle: CSSProperties = {
        width: overviewWidth,
        minWidth: overviewWidth,
        flexGrow: 0,
        display: 'flex',
        flexDirection: 'column',
        gap: 10,
    };

    const imgWidth = overviewWidth;
    const imgHeight = overviewWidth * 2/3;

    const onImageUpload = (dataURI: string) => {
        const newRecipe: Recipe = {...recipe};
        newRecipe.image_url = dataURI;
        setRecipe(newRecipe);
    }

    return (
        <div style={overviewStyle}>
            <ImageUpload width={imgWidth} height={imgHeight} onImage={onImageUpload} maxArea={1500 * 1500}>
                {recipe.image_url !== '' &&
                <CentererdImage
                    src={recipe.image_url}
                    alt={recipe.title}
                    width={imgWidth}
                    height={imgHeight}
                    style={{borderRadius: 16}}/>}
            </ImageUpload>
            <TextField fullWidth name="Title">
                <Label>Title</Label>
                <Input placeholder="Recipe Title" value={recipe.title} onChange={(e) => {
                    const newRecipe: Recipe = {...recipe};
                    newRecipe.title = e.currentTarget.value;
                    setRecipe(newRecipe);
                }}/>
            </TextField>
            <div style={{display: 'flex', flexDirection: 'column', gap: 2}}>
                <Label>Description</Label>
                <TextArea fullWidth rows={6} placeholder="Recipe description" value={recipe.description} onChange={(e) => {
                    const newRecipe: Recipe = {...recipe};
                    newRecipe.description = e.currentTarget.value;
                    if (newRecipe.description.length > descriptionLimitChars) {
                        newRecipe.description = newRecipe.description.substring(0, descriptionLimitChars);
                    }
                    setRecipe(newRecipe);
                }}/>
                <Description>
                    Characters: {recipe.description.length} / {descriptionLimitChars}
                </Description>
            </div>

            <RecipeOverViewTagEdit recipe={recipe} setRecipe={setRecipe}/>
        </div>
    );
}

function RecipeOverViewTagEdit({recipe, setRecipe}: {recipe: Recipe, setRecipe: (r: Recipe) => void}) {
    const [tagOptions, setTagOptions] = useState<string[]>([]);

    const tagSelectExpanded = () => {
        if (tagOptions.length > 0) return;

        listTags().then((tags) => {
            setTagOptions(tags);
        }).catch((reason: RecipeAPIError) => {
            toast.danger("Failed to Retrieve Tags", {
                description: reason.message,
            });
        });
    };

    const tagContainerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        gap: 5,
    };

    return (
        <div style={tagContainerStyle}>
            {recipe.tags?.map((tag, idx) => {
                return <RecipeTag tag={tag.name} special={tag.id === 0} removable onClick={() => {
                    const newRecipe: Recipe = {...recipe};
                    recipe.tags.splice(idx, 1);
                    setRecipe(newRecipe);
                }}/>
            })}

            <PopupTypeSelect
                allowCreate
                onExpanded={tagSelectExpanded}
                onSelect={(v) => {
                    if (recipe.tags?.find((existing) => existing.name === v.key)) {
                        return; // Don't add duplicate tags
                    }

                    const id = v.value ? v.value : 0;
                    const newRecipe: Recipe = {...recipe};
                    if (newRecipe.tags === null) {
                        newRecipe.tags = [];
                    }
                    newRecipe.tags.push({name: v.key, id: id, user_id: 0});
                    setRecipe(newRecipe);
                }}
                options={tagOptions.map((v: string, idx: number): PopupTypeSelectOption<number> => {
                    return {key: v, value: idx};
                })}>
                <RecipeTag tag={recipe.tags?.length > 0 ? "" : "Add Tag"} add/>
            </PopupTypeSelect>
        </div>);
}