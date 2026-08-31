import { useLayoutEffect, useRef, useState, type CSSProperties } from "react";
import type { RecipeIngredient, RecipeStep } from "../../apiTypes";
import { fullTimeLabel, quantityLabel } from "../../shared/recipeHelpers";
import { Typography } from "@heroui/react";


export function RecipeStepDescriptionText({step, onEdit, style}: {step: RecipeStep, onEdit?: (v: string) => void, style?: CSSProperties}) {
    const fullText = step.body_text.length > 0 ? step.body_text : ' ';
    const bodyParagraphs = fullText.split('\n');
    const containerRef = useRef<HTMLDivElement>(null);
    const caretRef = useRef<number | null>(null);

    const textContainerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
        gap: 5,
        // Substitute &nbsp with space
        whiteSpace: 'pre-wrap',
        ...style,
    };

    // Restore the caret to position prior to DOM changes
    useLayoutEffect(() => {
        const root = containerRef.current;
        const offset = caretRef.current;
        caretRef.current = null;

        if (root == null || offset == null) return;

        if (document.activeElement !== root) root.focus();
        setCaretOffset(root, offset);
    });

    const onChange = (e: React.InputEvent<HTMLDivElement>) => {
        const root = e.currentTarget;
        const edit = htmlToDescription(root, fullText);

        // Adjust for link removal
        let caretPos = adjustedCaretOffset(root, edit.linkFixups) ?? 0;
        let text = edit.text;

        // Adjust for placeholder space
        if (step.body_text.length === 0 && edit.text.length > 0) {
            if (edit.text.at(edit.text.length - 1) === ' ') {
                text = edit.text.substring(0, edit.text.length - 1);
            } else if (edit.text.at(0) === ' ') {
                caretPos -= 1;
                text = text.substring(1);
            }
        }

        caretRef.current = caretPos;
        for (const fixup of edit.linkFixups) {
            applyLinkFixup(fixup);
        }

        onEdit?.(text);
    }

    let paragraphStart = 0;
    const paragraphs = [];
    for (const idx in bodyParagraphs) {
        const text = bodyParagraphs[idx];
        paragraphs.push(<RecipeStepDescriptionTextParagraph key={idx} paragraph={text} paragraphStart={paragraphStart} step={step}/>);
        paragraphStart += text.length + 1; // + 1 for the '\n' that split() consumed
    }

    // descriptionStructure key forces react to drop entire DOM tree when a major change happens, for example removing a paragraph
    return (
        <div key={descriptionStructure(bodyParagraphs)} ref={containerRef} style={textContainerStyle} contentEditable={onEdit != undefined} onInput={onChange} suppressContentEditableWarning={true}>
            {paragraphs}
        </div>
    );
}

type textStop = {
    node: Text;
    start: number;
};

type span = {
    start: number;
    end: number;
};

type scannedText = {
    // In document order.
    stops: textStop[];
    elements: Map<HTMLElement, span>;
};

// scanText walks the editable content and assigns every text node and element a position
function scanText(root: HTMLElement): scannedText {
    const stops: textStop[] = [];
    const elements = new Map<HTMLElement, span>();
    let offset = 0;

    // separated: this element's children are the paragraphs of the description and are
    // joined with a newline. Only true for the container itself.
    const walk = (elem: HTMLElement, separated: boolean) => {
        const start = offset;
        let seenChild = false;

        for (const child of elem.childNodes) {
            if (!(child instanceof Text) && !(child instanceof HTMLElement)) continue;

            // Filler takes up no space in the description, so it must take up none here.
            if (child instanceof HTMLElement && isPlaceholderBr(child)) {
                elements.set(child, {start: offset, end: offset});
                continue;
            }

            if (separated && seenChild) offset += 1;
            seenChild = true;

            if (child instanceof Text) {
                stops.push({node: child, start: offset});
                offset += child.data.length;
            } else if (child.tagName === 'BR') {
                elements.set(child, {start: offset, end: offset + 1});
                offset += 1;
            } else {
                walk(child, false);
            }
        }

        elements.set(elem, {start: start, end: offset});
    };

    walk(root, true);
    return {stops: stops, elements: elements};
}

// getCaretOffset returns the caret position in the coordinates scanText assigns, or null
function getCaretOffset(root: HTMLElement, scan: scannedText): number | null {
    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0) return null;

    const range = selection.getRangeAt(0);
    if (!root.contains(range.startContainer)) return null;

    const node = range.startContainer;
    if (node instanceof Text) {
        const stop = scan.stops.find((s) => s.node === node);
        return stop == undefined ? null : stop.start + Math.min(range.startOffset, node.data.length);
    }

    if (!(node instanceof HTMLElement)) return null;

    // The caret can also be expressed as a position between an element's children, which
    // is where the browser leaves it when the text node it was in has just been removed.
    const child = node.childNodes[range.startOffset];
    if (child instanceof Text) {
        return scan.stops.find((s) => s.node === child)?.start ?? null;
    } else if (child instanceof HTMLElement) {
        return scan.elements.get(child)?.start ?? null;
    }

    // Past the last child.
    return scan.elements.get(node)?.end ?? null;
}

// setCaretOffset is the inverse of getCaretOffset.
function setCaretOffset(root: HTMLElement, offset: number) {
    const scan = scanText(root);
    const range = document.createRange();

    // An empty paragraph holds no text node to anchor to, so it has to be matched on the
    // element itself. This has to come first: the text node that follows an empty
    // paragraph would otherwise swallow the offset and pull the caret out of it.
    // An element is recorded once its children are, so the first match is the innermost
    // one; the caret belongs there rather than in an ancestor, so that text typed into it
    // lands inside the typography element and is styled like the rest.
    let empty: HTMLElement | undefined = undefined;
    for (const [elem, span] of scan.elements) {
        if (span.start === offset && span.end === offset) {
            empty = elem;
            break;
        }
    }

    const stop = scan.stops.find((s) => offset <= s.start + s.node.data.length);
    if (empty != undefined) {
        range.selectNodeContents(empty);
        range.collapse(true);
    } else if (stop != undefined) {
        range.setStart(stop.node, Math.max(0, offset - stop.start));
        range.collapse(true);
    } else {
        // Ran off the end of the content, e.g. characters were dropped by the re-render.
        range.selectNodeContents(root);
        range.collapse(false);
    }

    const selection = window.getSelection();
    selection?.removeAllRanges();
    selection?.addRange(range);
}

// A link whose rendered text no longer matches the value it stands for.
type linkFixup = {
    elem: HTMLElement;
    // The text the link will show after the fixup, or undefined if it is being dropped.
    displayText: string | undefined;
};

type descriptionEdit = {
    text: string;
    // In document order, so caret corrections can be accumulated in one pass.
    linkFixups: linkFixup[];
};

function htmlToDescription(root: HTMLElement, originalText: string): descriptionEdit {
    const linkFixups: linkFixup[] = [];

    // Each element child of the container is one paragraph, i.e. one '\n'-separated run of
    // the description. A bare text node can be left behind at this level when an edit
    // removes a whole paragraph.
    const paragraphs: string[] = [];
    for (const child of root.childNodes) {
        if (child instanceof Text) {
            paragraphs.push(normalizeText(child.data));
        } else if (child instanceof HTMLElement) {
            if (isPlaceholderBr(child)) continue;
            paragraphs.push(nodeToDescription(child, originalText, linkFixups));
        }
    }

    return {text: paragraphs.join('\n'), linkFixups: linkFixups};
}

// isPlaceholderBr reports whether a <br> is browser-inserted filler to give an empty line
// a caret position, rather than a line break the user typed. Filler is always last.
function isPlaceholderBr(elem: HTMLElement): boolean {
    if (elem.tagName !== 'BR') return false;

    for (let next = elem.nextSibling; next != null; next = next.nextSibling) {
        if (next instanceof HTMLElement) return false;
        if (next instanceof Text && next.data.length > 0) return false;
    }

    return true;
}

function nodeToDescription(elem: HTMLElement, originalText: string, linkFixups: linkFixup[]): string {
    if (elem.tagName === 'BR') {
        return isPlaceholderBr(elem) ? '' : '\n';
    }

    const dispText = elem.getAttribute('data-displaytext');
    if (dispText !== null) {
        return linkToDescription(elem, dispText, originalText, linkFixups);
    }

    let combined = '';
    for (const child of elem.childNodes) {
        // Read the characters off the text node rather than the parent's innerHTML, which
        // hands back entity escapes (&nbsp;, &amp;) as literal text.
        if (child instanceof Text) {
            combined += normalizeText(child.data);
        } else if (child instanceof HTMLElement) {
            combined += nodeToDescription(child, originalText, linkFixups);
        }
    }

    return combined;
}

function linkToDescription(elem: HTMLElement, dispText: string, originalText: string, linkFixups: linkFixup[]): string {
    const rendered = normalizeText(elem.textContent ?? '');
    if (rendered.length < dispText.length) {
        // Any part of the link removed removes all of it.
        linkFixups.push({elem: elem, displayText: undefined});
        return '';
    }
    if (rendered !== dispText) {
        linkFixups.push({elem: elem, displayText: dispText});
    }

    const start = parseInt(elem.getAttribute('data-start') ?? '');
    const length = parseInt(elem.getAttribute('data-length') ?? '');
    if (isNaN(start) || isNaN(length)) {
        console.error("link element is missing its position in the description");
        return '';
    }

    return originalText.slice(start, start + length);
}

// applyLinkFixup restores the display text of a link the user typed into. A dropped link
// needs nothing: it is gone from the description, so the re-render removes the element.
function applyLinkFixup(fixup: linkFixup) {
    if (fixup.displayText === undefined) return;

    // Write to the element holding the text rather than the link wrapper, which would
    // throw away the typography element and its classes.
    const textElem = fixup.elem.firstElementChild ?? fixup.elem;
    textElem.textContent = fixup.displayText;
}

// normalizeText undoes the non-breaking spaces contenteditable uses to stop runs of
// whitespace collapsing; the description should only ever hold ordinary spaces.
function normalizeText(text: string): string {
    return text.replaceAll(' ', ' ');
}

// adjustedCaretOffset measures the caret and moves it into the coordinates of the text
// that is about to be rendered, which is shorter than what is in the DOM wherever a link
// was dropped or trimmed back.
function adjustedCaretOffset(root: HTMLElement, linkFixups: linkFixup[]): number | null {
    const scan = scanText(root);
    const caret = getCaretOffset(root, scan);
    if (caret == null) return null;

    let shift = 0;
    for (const fixup of linkFixups) {
        const start = scan.elements.get(fixup.elem)?.start;
        if (start == undefined) continue;

        const oldLength = normalizeText(fixup.elem.textContent ?? '').length;
        const newLength = fixup.displayText?.length ?? 0;

        if (caret <= start) {
            // The rest of the fixups are past the caret and cannot move it.
            break;
        } else if (caret >= start + oldLength) {
            shift += newLength - oldLength;
        } else {
            // The caret sat inside the link that is being rewritten, so there is no
            // character left to anchor it to. Put it at the end of whatever remains.
            return start + shift + newLength;
        }
    }

    return caret + shift;
}

type contentPart = {
    str: string;
    idx: number;
};

function getContentParts(str: string): contentPart[] {
    const contentPrepped = str.replaceAll('{{', '||{{').replaceAll('}}', '}}||');
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

function RecipeStepDescriptionTextParagraph({paragraph, paragraphStart, step}: {paragraph: string, paragraphStart: number, step: RecipeStep}) {
    const ingredientPrefix = '{{ingr-';
    const timePrefix = '{{time-';

    // Offsets are recorded against the whole description, since that is what the link
    // elements are read back against, not against this paragraph alone.
    const contentParts = getContentParts(paragraph).map((part) => ({str: part.str, idx: paragraphStart + part.idx}));

    const isIngredientIdx = (chunk: string): number | undefined => {
        if (chunk.startsWith(ingredientPrefix) && chunk.endsWith("}}")) {
            return parseInt(chunk.substring(ingredientPrefix.length, chunk.length - 2));
        }
        return undefined;
    }

    const isTimeIdx = (chunk: string): number | undefined => {
        if (chunk.startsWith(timePrefix) && chunk.endsWith("}}")) {
            return parseInt(chunk.substring(timePrefix.length, chunk.length - 2));
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
                    return <RecipeStepInlineLink key={idx} text={stepInlineIngredientText(ingrIdx, step.ingredients)} pos={{start: chunk.idx, length: chunk.str.length}}/>;
                }
                const timeIdx = isTimeIdx(chunk.str);
                if (timeIdx != undefined) {
                    return <RecipeStepInlineLink key={idx} text={fullTimeLabel(step.times[timeIdx])} pos={{start: chunk.idx, length: chunk.str.length}}/>;
                }

                return (<span key={idx} className="typography typography--align-start typography--color-default typography--body-sm">
                    {chunk.str}
                </span>);
            })}
        </span>
    );
}

type textPos = {
    start: number;
    length: number;
};

function RecipeStepInlineLink({text, pos}: {text: string, pos: textPos}) {
    const [hovered, setHovered] = useState<boolean>(false);

    const typographyStyle: CSSProperties = {
        display: 'inline',
        textDecoration: 'underline',
        textDecorationThickness: hovered ? 2 : 1.5,
        cursor: 'pointer',
    };

    return (
        <div style={{display: 'inline'}} data-start={pos.start} data-length={pos.length} data-displaytext={text}>
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

function stepInlineIngredientText(idx: number, ingredients: RecipeIngredient[]): string {
    if (idx < 0 || idx > ingredients.length) {
        return "??ingredient??";
    }

    const target = ingredients[idx];

    // If another ingredient of the same name exists, must specify quantity too
    for (let i = 0; i < ingredients.length; i++) {
        if (i == idx) continue;

        if (ingredients[i].ingredient.name === target.ingredient.name) {
            return quantityLabel(target.quantity, target.unit) + ' ' + target.ingredient.name;
        }
    }

    return target.ingredient.name;
}