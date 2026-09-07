import { useMemo, useState, type CSSProperties } from "react";
import type { RecipeStep } from "../../apiTypes";
import { splitStepBody } from "../../shared/recipeHelpers";


export function RecipeStepEditDescriptionText({step, setStep}: {step: RecipeStep, setStep: (s: RecipeStep) => void}) {
    const [currentText, setCurrentText] = useState(step.body_text);
    const [prevStepText, setPrevStepText] = useState(step.body_text);

    if (step.body_text !== prevStepText) {
        setPrevStepText(step.body_text);
        setCurrentText(step.body_text);
    }

    const sharedTextProperties: CSSProperties = {
        width: '100%',
        whiteSpace: 'pre-wrap',
        paddingTop: 8,
        paddingBottom: 8,
        paddingLeft: 12,
        paddingRight: 12,
    };

    const textContainerStyle: CSSProperties = {
        minHeight: 100,
        backgroundColor: 'var(--default)',
        borderRadius: 12,
        ...sharedTextProperties,
    }

    const onFocusLoss = (e: React.FocusEvent<HTMLDivElement, Element>) => {
        const newStep: RecipeStep = {...step};
        newStep.body_text = getDivInnerText(e.currentTarget);
        setStep(newStep);
    }

    return (
        <div style={{position: 'relative', width: '100%'}}>
            <div contentEditable={true} style={textContainerStyle} className="input no-box-unfocused" onBlur={onFocusLoss} onInput={(e) => setCurrentText(getDivInnerText(e.currentTarget))}>
                {step.body_text}
            </div>
            <OverlayContent text={currentText} sharedStyle={sharedTextProperties}/>
        </div>);
}

function OverlayContent({text, sharedStyle}: {text: string, sharedStyle: CSSProperties}) {
    const overlayStyle: CSSProperties = {
        position: 'absolute',
        left: 0,
        background: 'none',
        top: 0,
        color: 'transparent',
        pointerEvents: 'none',
        ...sharedStyle,
    };

    const split = useMemo(() => {
        return splitStepBody(text);
    }, [text]);

    return (<div style={overlayStyle} className="input no-box-unfocused">
        {split.map((v, i) => {
            let linkStyle: CSSProperties | undefined = undefined;
            if (v.startsWith("{i:")) {
                linkStyle = {
                    backgroundColor: 'var(--accent-soft)',
                };
            } else if (v.startsWith("{t:")) {
                linkStyle = {
                    backgroundColor: 'var(--success-soft)',
                };
            }

            return <span style={linkStyle} key={i}>{v}</span>;
        })}
    </div>);
}

const blockTags = new Set(['DIV', 'P', 'LI', 'BLOCKQUOTE', 'PRE', 'TR', 'H1', 'H2', 'H3', 'H4', 'H5', 'H6']);

// Browsers represent a line break in a contentEditable in several ways (a <br>, a nested block
// element, or a literal newline), so walk the tree and normalize them all to '\n'.
function getDivInnerText(div: HTMLDivElement): string {
    const lines: string[] = [];
    let line = '';
    let started = false;

    const breakLine = () => {
        lines.push(line);
        line = '';
    };

    const walk = (node: Node) => {
        node.childNodes.forEach((child) => {
            if (child.nodeType === Node.TEXT_NODE) {
                const parts = (child.nodeValue ?? '').split('\n');
                line += parts[0];
                for (let i = 1; i < parts.length; i++) {
                    breakLine();
                    line = parts[i];
                }
                started = started || parts.length > 1 || parts[0] !== '';
                return;
            }

            if (child.nodeType !== Node.ELEMENT_NODE) {
                return;
            }

            const el = child as HTMLElement;

            if (el.tagName === 'BR') {
                // A trailing <br> is filler the browser adds to keep an empty block visible,
                // not a break the user typed.
                if (el !== node.lastChild) {
                    breakLine();
                }
                started = true;
                return;
            }

            if (blockTags.has(el.tagName)) {
                if (started) {
                    breakLine();
                }
                started = true;
            }

            walk(el);
        });
    };

    walk(div);
    lines.push(line);

    // contentEditable stores typed spaces as non-breaking spaces.
    return lines.join('\n').replace(/\u00a0/g, ' ');
}
