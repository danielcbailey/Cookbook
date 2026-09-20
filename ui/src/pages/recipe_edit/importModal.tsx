import { Camera, Sparkles } from "@gravity-ui/icons";
import { Button, CloseIcon, Input, Label, Modal, Spinner, toast, Typography } from "@heroui/react";
import { useMemo, useState, type CSSProperties } from "react";
import { importRecipeFromPhotos, importRecipeFromWeb } from "../../recipeAPI";
import type { Recipe } from "../../apiTypes";
import { CentererdImage } from "../../shared/recipeCard";
import { ImageUploadMultiple } from "../../shared/imageUpload";


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
                        {mode === 'web' ? <ImportWebForm onRecipeImport={onRecipeImport}/> : <ImportPhotoForm onRecipeImport={onRecipeImport}/>}
                    </Modal.Body>
                </Modal.Dialog>}
            </Modal.Container>
        </Modal.Backdrop>
    );
}

const formStyle: React.CSSProperties = {
    display: 'flex',
    gap: 5,
    flexDirection: 'column',
};

function ImportWebForm({onRecipeImport}: {onRecipeImport?: (recipe: Recipe) => void}) {
    const [url, setUrl] = useState<string>('');
    const [loading, setLoading] = useState<boolean>(false);

    const onSubmit = () => {
        setLoading(true);

        let adjustedURL = url;
        if (!url.includes("://")) {
            adjustedURL = "https://" + url;
        }

        importRecipeFromWeb(adjustedURL).then((recipe: Recipe) => {
            setLoading(false);
            if (onRecipeImport) {
                onRecipeImport(recipe);
            }
        }).catch((reason) => {
            setLoading(false);
            toast.danger("Failed to Import Recipe", {
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

function ImportPhotoForm({onRecipeImport}: {onRecipeImport?: (recipe: Recipe) => void}) {
    const [images, setImages] = useState<File[]>();
    const [loading, setLoading] = useState<boolean>(false);

    const onSubmit = () => {
        setLoading(true);

        if (!images) {
            return;
        }

        importRecipeFromPhotos(images).then((recipe: Recipe) => {
            setLoading(false);
            if (onRecipeImport) {
                onRecipeImport(recipe);
            }
        }).catch((reason) => {
            setLoading(false);
            toast.danger("Failed to Import Recipe", {
                description: reason.message,
            });
        })
    }

    const imageURIs = useMemo<string[]>((): string[] => {
        const ret: string[] = [];
        if (!images) return ret;

        for (const img of images) {
            ret.push(URL.createObjectURL(img));
        }

        return ret;
    }, [images]);

    const gridStyle: CSSProperties = {
        display: 'flex',
        flexDirection: 'row',
        gap: 10,
        flexWrap: 'wrap',
        width: '100%',
    };

    const previewSize = 70;

    return (<div style={formStyle}>
        <div style={gridStyle}>
            {imageURIs.map((v, i) => {
                return <ImagePreview size={previewSize} dataURI={v} onClick={() => {
                    const newImages = images ? [...images] : [];
                    newImages.splice(i, 1);
                    setImages(newImages);
                }}/>;
            })}
            <ImageUploadMultiple width={previewSize} height={previewSize} onFiles={(files) => {
                const newImages = images ? [...images] : [];
                for (const f of files) {
                    newImages.push(f);
                }
                setImages(newImages);
            }}/>
        </div>

        {!loading && <Button style={{marginTop: 10}} isDisabled={!images || images.length === 0} variant="primary" onClick={onSubmit}>
            Import
        </Button>}
        {loading && <Spinner/>}
    </div>);
}

function ImagePreview({size, dataURI, onClick}: {size: number, dataURI: string, onClick: () => void}) {
    const [hovered, setHovered] = useState<boolean>(false);

    const containerStyle: CSSProperties = {
        width: size,
        height: size,
        position: 'relative',
        flexGrow: 0,
    };

    const overlayStyle: CSSProperties = {
        position: 'absolute',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        backgroundColor: '#00000040',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: 'pointer',
        borderRadius: 8,
        color: 'white',
    }

    return (<div style={containerStyle} onMouseEnter={() => setHovered(true)} onMouseLeave={() => setHovered(false)} onClick={onClick}>
        <CentererdImage src={dataURI} width={size} height={size}/>
        {hovered && <div style={overlayStyle}>
            <CloseIcon/>
        </div>}
    </div>);
}