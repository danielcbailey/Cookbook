import { Button, IconChevronDown, IconChevronRight, IconPlus, Typography } from '@heroui/react';
import { useState, type CSSProperties } from 'react'
import {Star} from '@gravity-ui/icons';
import { useNavigate } from 'react-router-dom';

// {tree} : {tree: Record<string, string[]>}
export function RecipeSideBrowser({focusValue, focusRequest} : {focusValue?: string, focusRequest?: (value: string) => void}) {
    const browserStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
        gap: 10,
        width: 200,
        paddingLeft: 20,
        paddingTop: 5,
    };

    const navigate = useNavigate();

    const onCreateRecipeClick = () => {
        navigate('/recipes/create');
    }

    return (
        <div style={browserStyle}>
            <Button variant="primary" className="font-semibold" style={{width: '100%'}} onClick={onCreateRecipeClick}>
                <IconPlus style={{width: 16, height: 16}}/>
                Create Recipe
            </Button>
            <RecipeTreeItemParent name="Suggested" icon={<Star style={{width: 16, height: 16}}/>} children={[]} focusedValue={focusValue} focusRequest={focusRequest} />
            <RecipeTreeItemParent name="Proteins" children={['Chicken', 'Beef', 'Pork', 'Fish']} focusedValue={focusValue} focusRequest={focusRequest} />
        </div>
    );
}

type RecipeTreeItemParentProps = {
    name: string;
    children: string[];
    icon?: React.ReactNode;
    focusedValue?: string;
    focusRequest?: (value: string) => void;
};

function RecipeTreeItemParent({name, children, icon, focusedValue, focusRequest} : RecipeTreeItemParentProps) {
    const holdsFocus = !!focusedValue && (focusedValue === name || children.includes(focusedValue));

    const [expanded, setExpanded] = useState(holdsFocus);

    // Expand when the focus moves onto this item or one of its children, while
    // still letting a click toggle it back.
    const [prevFocusedValue, setPrevFocusedValue] = useState(focusedValue);
    if (focusedValue !== prevFocusedValue) {
        setPrevFocusedValue(focusedValue);
        if (holdsFocus && children.length > 0) {
            setExpanded(true);
        }
    }

    let renderedIcon = icon;
    if (children.length > 0 && !icon) {
        renderedIcon = expanded ? <IconChevronDown/> : <IconChevronRight/>;
    }

    const clickHandler = () => {
        if (focusRequest && !expanded) {
            focusRequest(name);
        }

        if (children.length > 0) {
            setExpanded(!expanded);
        }
    }

    const childrenContainerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
        gap: 1,
    }

    return (
        <div style={childrenContainerStyle}>
            <RecipeTreeItem
                name={name}
                isChild={false}
                icon={renderedIcon}
                onClick={clickHandler}
                focused={name === focusedValue}
                />

            {expanded && children.map((childName) => (
                <RecipeTreeItem
                    key={childName}
                    name={childName}
                    isChild={true}
                    focused={childName === focusedValue}
                    onClick={() => focusRequest && focusRequest(childName)}
                />
            ))}
        </div>
    );
}

function RecipeTreeItem({name, isChild, icon, focused, onClick} : {name: string, isChild: boolean, icon?: React.ReactNode, focused?: boolean, onClick?: () => void}) {
    const [hovered, setHovered] = useState(false);

    const itemStyle: CSSProperties = {
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        cursor: 'pointer',
        marginLeft: !icon ? 27 : 0,
        paddingLeft: 5,
        paddingTop: isChild ? 2 : 5,
        paddingBottom: isChild ? 2 : 5,
        paddingRight: 5,
        backgroundColor: focused ? 'var(--color-background-secondary)' : undefined,
        borderRadius: focused ? 99 : undefined,
    }

    const textType = isChild ? 'body-sm' : 'body';

    return (
        <div 
            style={itemStyle}
            className="no-select"
            onMouseEnter={() => setHovered(true)}
            onMouseLeave={() => setHovered(false)}
            onClick={onClick}>

            {icon}
            <Typography type={textType} weight="semibold" color={hovered || focused ? 'default' : 'muted'}>
                {name}
            </Typography>
        </div>
    );
}