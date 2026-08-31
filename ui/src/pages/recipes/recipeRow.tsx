import { useLayoutEffect, useRef, useState, type CSSProperties } from 'react'
import { RecipeCard } from '../../shared/recipeCard';
import type { RecipeListItem } from '../../apiTypes';
import { Button, IconChevronLeft, IconChevronRight, Typography } from '@heroui/react';
import { useNavigate } from 'react-router-dom';
import { capitalizeWords } from '../../helpers';

const cardWidth = 180;
const cardGap = 10;

export function RecipeListRow({recipes, label}: {recipes: RecipeListItem[], label: string}) {
    const wrapperRef = useRef<HTMLDivElement>(null);
    const [pageSize, setPageSize] = useState(0);
    // Setter comes back when the row gets prev/next controls
    const [scrollIdx, setScrollIdx] = useState(0);
    const navigate = useNavigate();

    // The wrapper stretches to the available width, so measuring it tells us how many
    // cards fit. The row itself is clipped, so overflowing cards never widen the parent.
    useLayoutEffect(() => {
        const wrapper = wrapperRef.current;
        if (!wrapper) {
            return;
        }

        const measure = () => {
            // n cards occupy n*cardWidth + (n-1)*cardGap
            const fits = Math.floor((wrapper.clientWidth + cardGap) / (cardWidth + cardGap));
            setPageSize(Math.max(fits, 1));
        };

        measure();

        const observer = new ResizeObserver(measure);
        observer.observe(wrapper);
        window.addEventListener('resize', measure);

        return () => {
            observer.disconnect();
            window.removeEventListener('resize', measure);
        };
    }, []);

    const scrollFn = (delta: number) => {
        let newIdx = scrollIdx + delta;
        newIdx = Math.min(Math.max(recipes.length - pageSize, 0), Math.max(newIdx, 0));
        setScrollIdx(newIdx);
    }

    const rowStyle: CSSProperties = {
        position: 'relative',
        display: 'flex',
        flexDirection: 'row',
        gap: cardGap,
        minWidth: 0,
        padding: 5,
        overflow: 'hidden',
    };

    const rowWrapperStyle: CSSProperties = {
        display: 'flex',
        gap: 10,
        flexDirection: 'column',
        minWidth: 0,
    }

    // Clamped at render rather than synced through an effect, so a resize that grows the
    // page size can't leave a partially empty row for a frame
    const startIdx = Math.min(scrollIdx, Math.max(recipes.length - pageSize, 0));
    const visible = recipes.slice(startIdx, startIdx + pageSize);

    return (
        <div style={rowWrapperStyle} ref={wrapperRef}>
            <RecipeListRowHeader label={label}/>
            <div style={rowStyle}>
                {visible.map((recipe: RecipeListItem, idx: number) => {
                    return <RecipeCard
                                key={recipe.id + ':' + (startIdx + idx)}
                                recipe={recipe}
                                width={cardWidth}
                                onClick={() => navigate('/recipes/view/' + recipe.id.toString())}/>
                })}

                <RecipeListRowOverflowMarker onClick={() => scrollFn(-3)} show={scrollIdx > 0} toRight={false}/>
                <RecipeListRowOverflowMarker onClick={() => scrollFn(3)} show={scrollIdx + pageSize < recipes.length} toRight={true}/>
            </div>
        </div>
    );
}

function RecipeListRowOverflowMarker({toRight, show, onClick}: {toRight: boolean, show: boolean, onClick: () => void}) {
    const containerStyle: CSSProperties = {
        position: 'absolute',
        top: 0,
        left: toRight ? undefined : 0,
        right: toRight ? 0 : undefined,
        width: cardWidth / 2,
        height: '100%',
        display: show ? 'flex' : 'none',
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: toRight ? 'flex-end' : 'flex-start',
        background: 'linear-gradient(' + (toRight ? '90' : '270') + 'deg, rgba(0, 0, 0, 0.00) 0.04%, var(--background) 99.96%)',
    };

    return (
        <div style={containerStyle}>
            <Button isIconOnly variant="tertiary" onPress={onClick}>
                {toRight ? <IconChevronRight/> : <IconChevronLeft/>}
            </Button>
        </div>
    );
}

function RecipeListRowHeader({label}: {label: string}) {
    // The whitespace will become a chevron icon
    const parts = label.replaceAll('/', '/ /').split('/');

    const containerStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        gap: 4,
        alignItems: 'center',
    };

    return (
        <div style={containerStyle}>
            {parts.map((part: string, idx: number) => {
                if (part === ' ') {
                    return <IconChevronRight key={idx} style={{marginTop: 4}}/>
                }

                return <Typography type='h3'>{capitalizeWords(part)}</Typography>
            })}
        </div>
    );
}