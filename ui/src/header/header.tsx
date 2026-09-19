import { useContext, useState, type CSSProperties } from 'react'
import { useNavigate } from 'react-router-dom';
import { Avatar, Button, IconSearch, Typography } from '@heroui/react'
import { UserContext } from '../contexts';
import { LogoWordmark } from '../shared/brand';

export function Header({loginVariant}: {loginVariant?: boolean}) {
    const user = useContext(UserContext);

    const headerStyle: CSSProperties = {
        width: '100%',
        display: 'flex',
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: loginVariant ? 'flex-start' : 'center',
        padding: 10,
        borderBottom: loginVariant ? undefined : '2px solid var(--border)',
    }

    return (
        <div style={headerStyle}>
            <LogoWordmark/>
            {!loginVariant && <NavButtons/>}
            {!loginVariant && <NavRight user={user && (user.first_name + ' ' + user.last_name)}/>}
        </div>
    );
}

const headerFlexStyle: CSSProperties = {
    display: 'flex',
    flexDirection: 'row',
    gap: 10,
    alignItems: 'center',
}

function NavButtons() {
    const navigate = useNavigate();
    const currPath = window.location.pathname;

    return (
        <div style={headerFlexStyle}>
            <NavButton
                selected={currPath.startsWith('/plan')}
                name="Plan"
                onClick={() => navigate('/plan')}
            />
            <NavButton
                selected={currPath.startsWith('/recipes')}
                name="Recipes"
                onClick={() => navigate('/recipes')}
            />
            <NavButton
                selected={currPath.startsWith('/pantry')}
                name="Pantry"
                onClick={() => navigate('/pantry')}
            />
        </div>
    );
}

function NavButton({selected, name, onClick}: {selected: boolean, name: string, onClick: () => void}) {
    const [hovered, setHovered] = useState(false);

    const divStyle: CSSProperties = {
        height: 24,
        borderBottom: selected ? '2px solid var(--accent)' : undefined,
        cursor: 'pointer',
    }

    return (
        <div 
            onMouseEnter={() => setHovered(true)}
            onMouseLeave={() => setHovered(false)}
            onClick={onClick}
            className="no-select"
            style={divStyle}>
            <Typography type="body" weight="semibold" color={selected || hovered ? 'default' : 'muted'}>
                {name}
            </Typography>
        </div>
    )
}

function NavRight({user}: {user: string | null}) {
    const navigate = useNavigate();

    const userInitials = user && user.split(' ').map(name => name[0]).join('').toUpperCase();

    return (
        <div style={headerFlexStyle}>
            <Button variant="tertiary">
                <IconSearch/>
                <Typography type="body" weight="medium">
                    Search
                </Typography>
            </Button>

            <div onClick={() => navigate('/profile')} className="no-select" style={{cursor: 'pointer'}}>
                <Avatar color="accent">
                    <Avatar.Fallback>
                        {userInitials}
                    </Avatar.Fallback>
                </Avatar>
            </div>
            
        </div>
    );
}