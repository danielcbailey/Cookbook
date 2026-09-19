import { Typography } from "@heroui/react";
import type { CSSProperties } from "react";
import iconUrl from '../assets/icon.png'

export function LogoWordmark({style}: {style?: CSSProperties}) {
    const wordmarkStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        gap: 10,
        alignItems: 'center',
        ...style,
    };

    return (
        <div style={wordmarkStyle}>
            <img src={iconUrl} style={{width: 40, height: 40}}/>
            <Typography.Heading level={2} weight="bold">Cookbook</Typography.Heading>
        </div>
    );
}