import { ArrowUpFromLine, Picture } from "@gravity-ui/icons";
import { useRef, useState, type CSSProperties } from "react";

// JPEG quality of the encoded image. 0.85 is the usual point where further
// quality buys far more bytes than it does visible detail.
const imageQuality = 0.85;

export function ImageUpload({children, width, height, onImage, maxArea}: {children?: React.ReactNode, width: number, height: number, onImage?: (dataURI: string) => void, maxArea?: number}) {
    const [hovered, setHovered] = useState<boolean>(false);
    const inputRef = useRef<HTMLInputElement>(null);

    const containerStyle: CSSProperties = {
        position: 'absolute',
        top: 0,
        left: 0,
        display: 'flex',
        flexDirection: 'column',
        gap: 5,
        alignItems: 'center',
        justifyContent: 'center',
        width: width,
        height: height,
        backgroundColor: 'var(--default)',
        borderRadius: 16,
        cursor: 'pointer',
        opacity: !children ? undefined : 0.5,
    };

    const showUpload = hovered || !children;

    const pickFile = () => {
        inputRef.current?.click();
    }

    const fileSelected = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.currentTarget.files?.[0];

        // Clearing the value lets the same file be picked again later: without
        // it the input holds the old selection and fires no change event.
        event.currentTarget.value = '';
        if (!file) return;

        try {
            const resized = await resizeImage(file, maxArea);
            onImage?.(imageToURI(resized));
        } catch (error) {
            console.error("Error reading image:", error);
        }
    }

    return (
        <div 
            style={{position: 'relative', width: width, height: height, minWidth: width}}
            onClick={pickFile}
            onMouseEnter={() => setHovered(true)}
            onMouseLeave={() => setHovered(false)}>
            {children}
            {showUpload && <div style={containerStyle}>
                <Picture style={{width: 30, height: 30}}/>
                Upload Image
            </div>}
            {/* The picker is opened by clicking the container, so the input
                itself stays hidden. Its own click must not bubble back up to
                the container handler, which would re-open the dialog. */}
            <input
                ref={inputRef}
                type="file"
                accept="image/png,image/jpeg,image/bmp,image/webp"
                style={{display: 'none'}}
                onClick={(e) => e.stopPropagation()}
                onChange={fileSelected}/>
        </div>
    );
}

export function ImageUploadMultiple({width, height, onImages, onFiles, maxAreaEach}: {width: number, height: number, onImages?: (dataURIs: string[]) => void, onFiles?: (files: File[]) => void, maxAreaEach?: number}) {
    const inputRef = useRef<HTMLInputElement>(null);

    const containerStyle: CSSProperties = {
        position: 'absolute',
        top: 0,
        left: 0,
        display: 'flex',
        flexDirection: 'column',
        gap: 5,
        alignItems: 'center',
        justifyContent: 'center',
        width: width,
        height: height,
        backgroundColor: 'var(--default)',
        borderRadius: 16,
        cursor: 'pointer',
    };

    const pickFiles = () => {
        inputRef.current?.click();
    }

    const filesSelected = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const files = Array.from(event.currentTarget.files ?? []);

        // Clearing the value lets the same files be picked again later: without
        // it the input holds the old selection and fires no change event.
        event.currentTarget.value = '';
        if (files.length === 0) return;

        if (onFiles) {
            onFiles(files);
            return;
        }

        try {
            // Resized one at a time, in the order they were picked. Decoding
            // them all at once would hold every full-resolution bitmap in
            // memory
            const dataURIs: string[] = [];
            for (const file of files) {
                dataURIs.push(imageToURI(await resizeImage(file, maxAreaEach)));
            }

            onImages?.(dataURIs);
        } catch (error) {
            console.error("Error reading images:", error);
        }
    }

    return (
        <div
            style={{position: 'relative', width: width, height: height, minWidth: width}}
            onClick={pickFiles}>
            {/* Nothing is rendered underneath, so the prompt always shows. */}
            <div style={containerStyle}>
                <ArrowUpFromLine/>
            </div>
            {/* The picker is opened by clicking the container, so the input
                itself stays hidden. Its own click must not bubble back up to
                the container handler, which would re-open the dialog. */}
            <input
                ref={inputRef}
                type="file"
                multiple
                accept="image/png,image/jpeg,image/bmp,image/webp"
                style={{display: 'none'}}
                onClick={(e) => e.stopPropagation()}
                onChange={filesSelected}/>
        </div>
    );
}

/**
 * Decodes an image file and scales it down so that its pixel area fits within
 * maxArea, preserving the aspect ratio. Images already under the limit are kept
 * at their native size. The result is a canvas ready to be encoded by
 * imageToURI.
 */
async function resizeImage(file: File, maxArea?: number): Promise<HTMLCanvasElement> {
    // Cameras record orientation in EXIF rather than in the pixels, so ask the
    // decoder to apply it — otherwise phone photos come out rotated.
    const source = await createImageBitmap(file, {imageOrientation: 'from-image'});
    const area = source.width * source.height;

    // Scaling area by s^2 means each side scales by s.
    const scale = maxArea && area > maxArea ? Math.sqrt(maxArea / area) : 1;
    const width = Math.max(1, Math.round(source.width * scale));
    const height = Math.max(1, Math.round(source.height * scale));

    // Letting the decoder resample is both faster and cleaner than a single
    // drawImage step, which aliases badly at large downscale factors.
    const scaled = scale === 1 ? source : await createImageBitmap(source, {
        resizeWidth: width,
        resizeHeight: height,
        resizeQuality: 'high',
    });

    try {
        const canvas = document.createElement('canvas');
        canvas.width = width;
        canvas.height = height;

        const context = canvas.getContext('2d');
        if (!context) {
            throw new Error("2d canvas context is unavailable");
        }

        context.drawImage(scaled, 0, 0);
        return canvas;
    } finally {
        if (scaled !== source) {
            scaled.close();
        }
        source.close();
    }
}

/**
 * Encodes a resized image as a data URI, the form the recipe API stores in
 * image_url. Recipe photos have no transparency to preserve, so JPEG is used
 * for its much smaller encoded size.
 */
function imageToURI(image: HTMLCanvasElement): string {
    return image.toDataURL('image/jpeg', imageQuality);
}