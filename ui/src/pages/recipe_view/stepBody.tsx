import { useMemo, useState, type CSSProperties } from "react";
import type { RecipeIngredient, RecipeStep } from "../../apiTypes";
import { fullTimeLabel, getStepIngredients, parseIngredientStepString, parseStepTimeString, quantityLabel } from "../../shared/recipeHelpers";
import { Typography } from "@heroui/react";


export function RecipeStepDescriptionText({step, style}: {step: RecipeStep, style?: CSSProperties}) {
    const fullText = step.body_text.length > 0 ? step.body_text : ' ';
    const bodyParagraphs = fullText.split('\n');

    const textContainerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
        gap: 5,
        // Substitute &nbsp with space
        whiteSpace: 'pre-wrap',
        ...style,
    };

    // eslint-disable-next-line react-hooks/exhaustive-deps
    const ingredients = useMemo(() => getStepIngredients(step), [step.body_text]);

    let paragraphStart = 0;
    const paragraphs = [];
    for (const idx in bodyParagraphs) {
        const text = bodyParagraphs[idx];
        paragraphs.push(<RecipeStepDescriptionTextParagraph key={idx} paragraph={text} paragraphStart={paragraphStart} ingredients={ingredients}/>);
        paragraphStart += text.length + 1; // + 1 for the '\n' that split() consumed
    }

    // descriptionStructure key forces react to drop entire DOM tree when a major change happens, for example removing a paragraph
    return (
        <div key={descriptionStructure(bodyParagraphs)} style={textContainerStyle}>
            {paragraphs}
        </div>
    );
}

type contentPart = {
    str: string;
    idx: number;
};

function getContentParts(str: string): contentPart[] {
    const contentPrepped = str.replaceAll('{', '||{').replaceAll('}', '}||');
    const contentPartsStrs = contentPrepped.split('||');

    const ret: contentPart[] = [];
    let idx = 0;
    for (const part of contentPartsStrs) {
        ret.push({str: part, idx: idx});
        idx += part.length;
    }

    return ret;
}

// descriptionStructure summarises the element structure
function descriptionStructure(paragraphs: string[]): string {
    return paragraphs.map((paragraph) => getContentParts(paragraph).length).join(',');
}

function RecipeStepDescriptionTextParagraph({paragraph, paragraphStart, ingredients}: {paragraph: string, paragraphStart: number, ingredients: RecipeIngredient[]}) {
    const ingredientPrefix = '{i:';
    const timePrefix = '{t:';

    // Offsets are recorded against the whole description, since that is what the link
    // elements are read back against, not against this paragraph alone.
    const contentParts = getContentParts(paragraph).map((part) => ({str: part.str, idx: paragraphStart + part.idx}));

    const isIngredientIdx = (chunk: string): number | undefined => {
        if (chunk.startsWith(ingredientPrefix) && chunk.endsWith("}")) {
            return parseInt(chunk.substring(ingredientPrefix.length, chunk.length - 1));
        }
        return undefined;
    }

    const isTimeIdx = (chunk: string): number | undefined => {
        if (chunk.startsWith(timePrefix) && chunk.endsWith("}")) {
            return parseInt(chunk.substring(timePrefix.length, chunk.length - 1));
        }
        return undefined;
    }

    // An empty paragraph has no text to give it a line box, so without a minimum height a
    // caret placed in one would be invisible and the paragraph itself unclickable.
    return (
        <span style={{minHeight: '1em'}}>
            {contentParts.map((chunk: contentPart, idx: number) => {
                const ingrIdx = isIngredientIdx(chunk.str);
                if (ingrIdx != undefined) {
                    return <RecipeStepInlineLink key={idx} text={stepInlineIngredientText(chunk.str, ingredients)}/>;
                }
                const timeIdx = isTimeIdx(chunk.str);
                if (timeIdx != undefined) {
                    return <RecipeStepInlineLink key={idx} text={fullTimeLabel(parseStepTimeString(chunk.str))}/>;
                }

                return (<span key={idx} className="typography typography--align-start typography--color-default typography--body-sm">
                    {chunk.str}
                </span>);
            })}
        </span>
    );
}

function RecipeStepInlineLink({text}: {text: string}) {
    const [hovered, setHovered] = useState<boolean>(false);

    const typographyStyle: CSSProperties = {
        display: 'inline',
        textDecoration: 'underline',
        textDecorationThickness: hovered ? 2 : 1.5,
        cursor: 'pointer',
    };

    return (
        <div style={{display: 'inline'}}>
            <Typography 
                style={typographyStyle}
                type="body-sm"
                weight="semibold"
                onMouseLeave={() => setHovered(false)}
                onMouseEnter={() => setHovered(true)}>

                {text}
            </Typography>
        </div>
        
    );
}

function stepInlineIngredientText(rawText: string, ingredients: RecipeIngredient[]): string {
    const target = parseIngredientStepString(rawText);
    if (!target) {
        return rawText;
    }

    // If another ingredient of the same name exists, must specify quantity too
    for (let i = 0; i < ingredients.length; i++) {
        const candidate = ingredients[i];
        if (target.ingredient.name === candidate.ingredient.name && target.quantity === candidate.quantity && target.unit === candidate.unit) {
            continue;
        }

        if (candidate.ingredient.name === target.ingredient.name) {
            return quantityLabel(target.quantity, target.unit) + ' ' + target.ingredient.name;
        }
    }

    return target.ingredient.name;
}