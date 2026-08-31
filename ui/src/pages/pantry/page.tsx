import type { CSSProperties } from 'react'
import { Typography } from '@heroui/react'
import { Header } from '../../header/header'

const contentStyle: CSSProperties = {
    padding: 10,
}

export function PantryPage() {
    return (
        <>
            <Header/>
            <div style={contentStyle}>
                <Typography.Heading level={2} weight="bold">Pantry</Typography.Heading>
            </div>
        </>
    );
}
