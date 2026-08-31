import { type CSSProperties } from 'react';
import { type RecipeListItem } from '../apiTypes';
import { CloseIcon, IconPlus, Typography } from '@heroui/react';
import {Flame, Clock} from '@gravity-ui/icons';
import { timeUnitLabel } from './recipeHelpers';

export function RecipeCard({recipe, onClick, width, compact} : {recipe: RecipeListItem, onClick?: () => void, width: number, compact?: boolean}) {
    const cardStyle: CSSProperties = {
        width: width,
        flexDirection: compact ? 'row' : 'column',
        display: 'flex',
        gap: compact ? 5 : 0,
        borderRadius: 16,
        overflow: 'hidden',
        backgroundColor: 'var(--surface)',
        padding: compact ? 5 : 0,
        cursor: onClick ? 'pointer' : 'default',
    };

    return (
        <div style={cardStyle} onClick={onClick} className="shadow-surface no-select">
            <CentererdImage height={compact ? 60 : width * 2/3} width={compact ? 60 : width} src={recipe.image_url} alt={recipe.title}/>
            <RecipeCardOverview recipe={recipe} compact={compact}/>
        </div>
    );
}

export function CentererdImage({height, width, src, alt, style} : {height: number, width: number, src: string, alt?: string, style?: CSSProperties}) {
    const frameStyle: CSSProperties = {
        height: height,
        width: width,
        minHeight: height,
        minWidth: width,
        flexShrink: 0,
        overflow: 'hidden',
        borderRadius: width === height ? 8 : 0,
    };

    const imageStyle: CSSProperties = {
        height: '100%',
        width: '100%',
        objectFit: 'cover',
        objectPosition: 'center',
        display: 'block',
    };

    return (
        <div style={frameStyle}>
            <img style={{...imageStyle, ...style}} src={src} alt={alt ?? ''} />
        </div>
    );
}

function RecipeCardOverview({recipe, compact} : {recipe: RecipeListItem, compact?: boolean}) {
    const overviewStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
        gap: compact ? 0 : 5,
        padding: compact ? 0 : 10,
    };

    const textType = compact ? 'body-xs' : 'body-sm';

    const time = recipe.time.end_time ? recipe.time.end_time : recipe.time.start_time;

    const timeLabel = (Math.round(time * 10)/10).toString() + ' ' + timeUnitLabel(recipe.time.unit);

    return (
        <div style={overviewStyle}>
            <Typography type={textType} weight="semibold">{recipe.title}</Typography>
            <CompletelyTruncatingFlexbox height={18} style={{justifyContent: 'space-between'}}>
                <RecipeCardStat label={timeLabel} icon={<Flame style={{width: 12, height: 12, color: 'var(--muted)'}}/>}/>
                <RecipeCardStat label={recipe.calories.toString() + ' kcal'} icon={<Clock style={{width: 12, height: 12, color: 'var(--muted)'}}/>}/>
            </CompletelyTruncatingFlexbox>
            <CompletelyTruncatingFlexbox height={18} style={{gap: 5}}>
                {recipe.tags && recipe.tags.map((tag) => (
                    <RecipeTag key={tag.name + ":" + tag.id} tag={tag.name} special={tag.id === 0}/>
                ))}
            </CompletelyTruncatingFlexbox>
        </div>
    );
}

function CompletelyTruncatingFlexbox({children, height, style} : {children?: React.ReactNode, height: number, style?: CSSProperties}) {
    const containerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        flexWrap: 'wrap',
    };

    const wrapperStyle: CSSProperties = {
        height: height,
        overflow: 'hidden',
        flexShrink: 0,
    };

    return (
        <div style={wrapperStyle}>
            <div style={{...containerStyle, ...style}}>
                {children}
            </div>
        </div>
    );
}

export function RecipeCardStat({label, icon} : {label: string, icon: React.ReactNode}) {
    const statStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        gap: 2,
        alignItems: 'center',
        padding: 2,
    };

    return (
        <div style={statStyle}>
            {icon}
            <Typography type="body-xs" weight="semibold" color="muted" style={{lineHeight: '14px'}}>{label}</Typography>
        </div>
    )
}

export function RecipeTag({tag, special, removable, add, onClick} : {tag: string, special?: boolean, removable?: boolean, add?: boolean, onClick?: () => void}) {
    const tagStyle: CSSProperties = {
        backgroundColor: special ? 'var(--color-warning-soft)' : 'var(--color-accent-soft)',
        borderRadius: 99,
        paddingLeft: 4,
        paddingRight: 4,
        paddingTop: 1,
        paddingBottom: 1,
        border: special ? '1px solid var(--color-warning-soft-hover)' : '1px solid var(--color-accent-soft-hover)',
        display: 'flex',
        flexDirection: 'row',
        gap: 2,
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: 17,
        cursor: onClick ? 'pointer' : undefined,
    };

    const textColor = special ? 'var(--color-warning-soft-foreground)' : 'var(--color-accent-soft-foreground)';

    const textStyle: CSSProperties = {
        color: textColor,
        lineHeight: '12px',
        paddingBottom: 2,
    };

    return (
        <div style={tagStyle} onClick={() => onClick && onClick()}>
            {add && <IconPlus style={{width: 10, height: 10}} color={textColor}/>}
            {tag !== '' && <Typography type="body-xs" weight="semibold" style={textStyle}>
                {tag}
            </Typography>}
            {removable && <CloseIcon style={{width: 10, height: 10}} color={textColor}/>}
        </div>
    );
}
