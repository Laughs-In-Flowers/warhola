## Changelog

### 0.0.5 (18.05.2018)

- rewrite of canvas to eliminate bilevel canvas/image distortion and create 
  a single structure and interface for package needs (as well as compatibility 
  with image.Image) 


### Warhola 0.0.4 (14.03.2018)

- cleanup & document 
- plugin abstraction to accomodate built in functionality
- added transform plugins: crop, resize, rotate, flip, shear, translate
- added adjustment plugins: brightness, gamma, contrast, hue, saturation 
- removed plugins to work on canvas issues


### Warhola 0.0.3 (25.11.2017)

- refactor & rewrite
- status command improvements and integration
- remove central core of Factory to functionality stored in context.Context
- move factory name and concept to canvas with slightly different implementation
- Image interface reduces dependence on draw.Image
- simple draw, clone, copy and paste functions with association to Canvas
- util consolidation to remove repeat code
- util/ctx package to abstract most common context.Context interaction
- util/xrr package to aggregate a common error
- continued refinement of text functions


### Warhola 0.0.2 (31.10.2017)

- function based Anchors


### Warhola 0.0.1 (31.10.2017)

- initialize with changelog & readme 
