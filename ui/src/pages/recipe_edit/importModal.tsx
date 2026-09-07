import { Camera, Sparkles } from "@gravity-ui/icons";
import { Button, Input, Label, Modal, Spinner, toast, Typography } from "@heroui/react";
import { useState } from "react";
import { importRecipeFromWeb } from "../../recipeAPI";
import type { Recipe } from "../../apiTypes";


export function ImportModal({mode, onClose, onRecipeImport}: {mode: 'web' | 'photos' | null, onClose: () => void, onRecipeImport: (recipe: Recipe) => void}) {

    const setIsOpen = (open: boolean) => {
        if (!open) {
            onClose();
        }
    }

    return (
        <Modal.Backdrop isOpen={mode != null} onOpenChange={setIsOpen}>
            <Modal.Container>
                {mode != null && <Modal.Dialog style={{maxWidth: 500, color: 'var(--foreground)'}}>
                    <Modal.Header>
                        <Modal.Icon className="bg-default text-foreground">
                            {mode === 'web' ? <Sparkles className="size-5"/> : <Camera className="size-5"/>}
                        </Modal.Icon>
                        <Modal.Heading>
                            <Typography type="h5" weight="semibold">
                                {mode === 'web' ? "Import from Web" : "Import from Photos"}
                            </Typography>
                        </Modal.Heading>
                    </Modal.Header>
                    <Modal.Body>
                        {mode === 'web' && <ImportWebForm onRecipeImport={onRecipeImport}/>}
                    </Modal.Body>
                </Modal.Dialog>}
            </Modal.Container>
        </Modal.Backdrop>
    );
}

function ImportWebForm({onRecipeImport}: {onRecipeImport?: (recipe: Recipe) => void}) {
    const [url, setUrl] = useState<string>('');
    const [loading, setLoading] = useState<boolean>(false);

    const formStyle: React.CSSProperties = {
        display: 'flex',
        gap: 5,
        flexDirection: 'column',
    };

    const onSubmit = () => {
        setLoading(true);
        importRecipeFromWeb(url).then((recipe: Recipe) => {
            setLoading(false);
            if (onRecipeImport) {
                onRecipeImport(recipe);
            }
        }).catch((reason) => {
            setLoading(false);
            toast.danger("Failed to Retrieve Ingredients", {
                description: reason.message,
            });
        })
    }

    const urlPattern = new RegExp('^(https?:\\/\\/)?'+ // protocol
        '((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.)+[a-z]{2,}|'+ // domain name
        '((\\d{1,3}\\.){3}\\d{1,3}))'+ // OR ip (v4) address
        '(\\:\\d+)?(\\/[-a-z\\d%_.~+]*)*'+ // port and path
        '(\\?[;&a-z\\d%_.~+=-]*)?'+ // query string
        '(\\#[-a-z\\d_]*)?$','i'); // fragment locator
    const error = url && !urlPattern.test(url) ? "Please enter a valid URL in the format https://example.com/recipe" : undefined;

    return (
        <div style={formStyle}>
            <Label htmlFor="input-url">URL</Label>
            <Input variant="secondary" id="input-url" placeholder="https://example.com/recipe" type="text" value={url} onChange={(e) => setUrl(e.target.value)}/>
            {error && <span style={{color: 'var(--danger)'}}>{error}</span>}
            {!loading && <Button style={{marginTop: 10}} isDisabled={url.trim() === '' || !!error} variant="primary" onClick={onSubmit}>
                Import
            </Button>}
            {loading && <Spinner/>}
        </div>
    );
}