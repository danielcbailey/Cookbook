import type { CSSProperties } from 'react'
import { Typography } from '@heroui/react'
import { Header } from '../../header/header'

const contentStyle: CSSProperties = {
    padding: 10,
}

export function PlanPage() {
    return (
        <>
            <Header/>
            <div style={contentStyle}>
                <Typography.Heading level={2} weight="bold">Plan</Typography.Heading>
            </div>
        </>
    );
}
