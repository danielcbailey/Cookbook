import { Autocomplete, EmptyState, Header, Input, ListBox, Popover, SearchField, useFilter } from "@heroui/react";
import { useState, type CSSProperties } from "react";
import { exactMatch, useSearchFilterName } from "../../shared/filter";
import { IngredientUnit, type Ingredient, type RecipeIngredient } from "../../apiTypes";
import { quantityUnitLabel, validUnits, type unitDetails } from "../../shared/recipeHelpers";

import {Check} from '@gravity-ui/icons';

export type IngredientPopupProps = {
    children: React.ReactNode;
    style?: CSSProperties;
    onExpanded?: (expanded: boolean) => void;
    onSelect?: (v: RecipeIngredient) => void;
    allowCreate?: boolean;
    value?: RecipeIngredient;
    options: Ingredient[]; // Displayed in order
    maxMatches?: number; // Stops searching once this many options match
};

const DEFAULT_MAX_MATCHES = 50;

type stepOption = 'ingredient' | 'unit' | 'quantity';

type IngredientPopupSelectorProps = {
    parentProps: IngredientPopupProps;
    workingIngredient: RecipeIngredient;
    setWorkingIngredient: (r: RecipeIngredient) => void;
    step: stepOption;
    setStep: (s: stepOption) => void;
    entry: string;
    setEntry: (e: string) => void;
    done: () => void;
};

const defaultWorkingIngredient: RecipeIngredient = {
    id: 0,
    quantity: 0,
    unit: IngredientUnit.Unknown,
    ingredient: {
        id: 0,
        user_id: 0,
        name: '',
        category: '',
    }
};

export function IngredientPopup(props: IngredientPopupProps) {
    const [isOpen, setIsOpen] = useState<boolean>(false);
    const [entry, setEntry] = useState<string>('');
    const [step, setStep] = useState<stepOption>('ingredient');
    const [workingIngredient, setWorkingIngredient] = useState<RecipeIngredient>(props.value ? props.value : defaultWorkingIngredient);

    function setExpanded(expanded: boolean) {
        setIsOpen(expanded);
        if (expanded) {
            setWorkingIngredient(props.value ? props.value : defaultWorkingIngredient);
            setStep('ingredient');
            setEntry('');
        }
        props.onExpanded?.(expanded);
    }

    function done() {
        props.onSelect?.(workingIngredient);
        setExpanded(false);
    }

    function setPopupStep(s: stepOption) {
        if (s === 'quantity') {
            setEntry(workingIngredient.quantity > 0 ? workingIngredient.quantity.toString() : '');
        } else {
            setEntry('');
        }

        setStep(s);
    }

    const childrenProps: IngredientPopupSelectorProps = {
        parentProps: props,
        workingIngredient: workingIngredient,
        setWorkingIngredient: setWorkingIngredient,
        step: step,
        setStep: setPopupStep,
        entry: entry,
        setEntry: setEntry,
        done: done,
    };

    const contentWrapperStyle: CSSProperties = {
        padding: 10,
    };

    return (
        <Popover isOpen={isOpen} onOpenChange={setExpanded}>
            <Popover.Trigger>
                {props.children}
            </Popover.Trigger>

            <Popover.Content placement={'bottom'}>
                <Popover.Arrow/>

                <div style={contentWrapperStyle}>
                    {step === 'ingredient' && <PopupSelectIngredient props={childrenProps}/>}
                    {step === 'unit' && <PopupSelectUnit props={childrenProps}/>}
                    {step === 'quantity' && <PopupEnterQuantity props={childrenProps}/>}
                </div>
                
            </Popover.Content>
        </Popover>
    );
}

const listBoxStyle: CSSProperties = {
        maxHeight: 300,
        overflow: 'scroll',
    };

function PopupSelectIngredient({props}: {props: IngredientPopupSelectorProps}) {
    const maxMatches = props.parentProps.maxMatches ?? DEFAULT_MAX_MATCHES;

    // Stop scanning at maxMatches so a long options list costs no more than a short one.
    const matches = useSearchFilterName(props.entry, props.parentProps.options, maxMatches);

    function hasExactMatch(v: string): boolean {
        for (const match of matches) {
            if (exactMatch(v, match.name)) {
                return true;
            }
        }

        return false;
    }

    const onSelect = (ingr: Ingredient) => {
        const newWorkingIngredient: RecipeIngredient = {...props.workingIngredient};
        newWorkingIngredient.ingredient = ingr;
        props.setWorkingIngredient(newWorkingIngredient);
        props.setStep('unit');
    }

    return (
        <Autocomplete.Filter inputValue={props.entry} onInputChange={props.setEntry}>
            <SearchField autoFocus name="search" variant="secondary">
            <SearchField.Group>
                <SearchField.SearchIcon />
                <SearchField.Input placeholder="Search..." />
                <SearchField.ClearButton />
            </SearchField.Group>
            </SearchField>

            <PopupHeader props={props}/>
            
            <ListBox
                selectionMode="single"
                disallowEmptySelection
                style={listBoxStyle}
                selectedKeys={props.workingIngredient.ingredient.name !== '' ? [props.workingIngredient.ingredient.name] : undefined}
                onSelectionChange={(keys) => {
                    if (keys === 'all') return;

                    let selected = props.parentProps.options.find((o) => keys.has(o.id));
                    if (keys.has(-1)) {
                        selected = {name: props.entry, id: 0, user_id: 0, category: ""};
                    }
                    if (selected) onSelect(selected);
                }}
                renderEmptyState={() => <EmptyState>No results found</EmptyState>}
            >
                {props.parentProps.allowCreate && props.entry !== '' && !hasExactMatch(props.entry) && <ListBox.Item id={-1} textValue={""}>
                    Create: {props.entry}
                    <ListBox.ItemIndicator />
                </ListBox.Item>}
                {matches.map((item) => (
                    <ListBox.Item key={item.id} id={item.id} textValue={item.name}>
                        {item.name}
                        <ListBox.ItemIndicator />
                    </ListBox.Item>
                ))}
            </ListBox>
        </Autocomplete.Filter>
    );
}

function PopupSelectUnit({props}: {props: IngredientPopupSelectorProps}) {
    const {contains} = useFilter({sensitivity: "base"});

    // Unknown is the "not picked yet" state rather than a unit that can be chosen.
    const units = Object.values(validUnits).filter((u: unitDetails) => u.id !== IngredientUnit.Unknown);
    units.sort((a: unitDetails, b: unitDetails) => a.singular.localeCompare(b.singular));

    return (
        <Autocomplete.Filter filter={contains}>
            <SearchField autoFocus name="search" variant="secondary">
            <SearchField.Group>
                <SearchField.SearchIcon />
                <SearchField.Input placeholder="Select Unit" />
                <SearchField.ClearButton />
            </SearchField.Group>
            </SearchField>

            <PopupHeader props={props}/>

            <ListBox 
                selectedKeys={props.workingIngredient.unit !== IngredientUnit.Unknown ? [props.workingIngredient.unit] : undefined}
                selectionMode="single"
                disallowEmptySelection
                style={listBoxStyle}
                onSelectionChange={(keys) => {
                    if (keys === 'all') return;

                    for (const unit of keys) {
                        const newWorkingIngredient: RecipeIngredient = {...props.workingIngredient};
                        newWorkingIngredient.unit = unit.toString() as IngredientUnit;

                        if (!validUnits[newWorkingIngredient.unit].usesQuantity) {
                            newWorkingIngredient.quantity = 1;
                        }

                        props.setWorkingIngredient(newWorkingIngredient);
                        props.setStep('quantity');
                        break;
                    }
                }}
                renderEmptyState={() => <EmptyState>No unit found</EmptyState>}>
                <ListBox.Section>
                    <Header>Weight</Header>
                    {units.map((u: unitDetails) => {
                        if (u.volume) {
                            return null;
                        }

                        const textValue = u.singular === '' ? u.id : u.singular;

                        return (<ListBox.Item key={u.id} id={u.id} textValue={textValue}>
                            {textValue}
                        </ListBox.Item>);
                    })}
                </ListBox.Section>
                <ListBox.Section>
                    <Header>Volume</Header>
                    {units.map((u: unitDetails) => {
                        if (!u.volume) {
                            return null;
                        }

                        const textValue = u.singular === '' ? u.id : u.singular;

                        return (<ListBox.Item key={u.id} id={u.id} textValue={textValue}>
                            {textValue}
                        </ListBox.Item>);
                    })}
                </ListBox.Section>
            </ListBox>
        </Autocomplete.Filter>
    );
}

function PopupEnterQuantity({props}: {props: IngredientPopupSelectorProps}) {

    const error = parseFloat(props.entry) > 0 ? null : 'Invalid quantity';

    const onSubmit = () => {
        if (error) {
            return;
        }

        if (props.workingIngredient.unit !== IngredientUnit.Unknown && props.workingIngredient.ingredient.name !== '') {
            props.done();
        }
    }

    const wrapperStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
    };

    return (<div style={wrapperStyle}>
        <Input 
            min={0}
            value={props.entry}
            onChange={(event) => {
                const value = event.target.value;
                props.setEntry(value);
                const floatValue = parseFloat(value);
                if (floatValue > 0) {
                    const newWorkingIngredient: RecipeIngredient = {...props.workingIngredient};
                    newWorkingIngredient.quantity = floatValue;
                    props.setWorkingIngredient(newWorkingIngredient);
                }
            }}
            autoFocus
            onKeyDown={(event) => {
                if (event.key !== 'Enter') {
                    return;
                }

                event.preventDefault();
                onSubmit();
            }}
            placeholder={"Quantity"}
            type={"number"}/>
        {error && <span style={{color: 'var(--danger)'}}>{error}</span>}
        <PopupHeader props={props}/>
    </div>);
}

function PopupHeader({props}: {props: IngredientPopupSelectorProps}) {
    const headerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        justifyContent: 'space-between',
        padding: 10,
        gap: 10,
    }

    const containerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        gap: 10,
    };

    const name = props.workingIngredient.ingredient.name;
    const unit = props.workingIngredient.unit;
    const qty = props.workingIngredient.quantity;

    const ready = qty !== 0 && unit !== IngredientUnit.Unknown && name !== '';

    const checkStyle: CSSProperties = {
        backgroundColor: ready ? 'var(--accent)' : 'var(--default)',
        color: ready ? 'var(--accent-foreground)' : 'var(--muted)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        borderRadius: 8,
        width: 21,
        paddingTop: 2,
        cursor: ready ? 'pointer' : 'not-allowed',
    };

    const unitText = quantityUnitLabel(unit).singular;

    return (
        <div style={headerStyle}>
            <div style={containerStyle}>
                <PopupHeaderButton
                    value={qty !== 0 ? qty.toString() : undefined}
                    placeholder={'Qty'}
                    onClick={() => props.setStep('quantity')}
                    focused={props.step === 'quantity'}/>
                <PopupHeaderButton
                    value={unit !== IngredientUnit.Unknown ? (unitText === '' ? unit : unitText) : undefined}
                    placeholder={'Unit'}
                    onClick={() => props.setStep('unit')}
                    focused={props.step === 'unit'}/>
                <PopupHeaderButton
                    value={name !== '' ? name : undefined}
                    placeholder={'Ingredient'}
                    onClick={() => props.setStep('ingredient')}
                    focused={props.step === 'ingredient'}/>
            </div>

            <div style={checkStyle} onClick={ready ? () => props.done() : undefined}>
                    <Check/>
                </div>
        </div>
    );
}

function PopupHeaderButton({value, placeholder, onClick, focused}: {value?: string, placeholder: string, onClick?: () => void, focused: boolean}) {
    const [hovered, setHovered] = useState(false);

    const filled = value != null;
    const label = filled ? value : placeholder;

    let color = 'var(--default)';
    let textColor = 'var(--default-foreground)';
    if (focused && hovered) {
        color = 'var(--color-accent-hover)';
        textColor = 'var(--accent-foreground)';
    } else if (focused) {
        color = 'var(--accent)';
        textColor = 'var(--accent-foreground)';
    } else if (filled && hovered) {
        color = 'var(--color-accent-soft-hover)';
        textColor = 'var(--color-accent-soft-foreground)';
    } else if (filled) {
        color = 'var(--color-accent-soft)';
        textColor = 'var(--color-accent-soft-foreground)';
    } else if (hovered) {
        color = 'var(--color-default-hover)';
    }

    const buttonStyle: CSSProperties = {
        backgroundColor: color,
        color: textColor,
        borderRadius: 8,
        paddingLeft: 8,
        paddingRight: 8,
        paddingBottom: 1,
        cursor: 'pointer',
    };

    return <span style={buttonStyle} onMouseEnter={() => setHovered(true)} onMouseLeave={() => setHovered(false)} onClick={() => onClick?.()}>{label}</span>;
}