## Changelog

### warhola 0.0.7 (04.12.2018)
- removal of imagick integration from main (by decision to not worry about
  integration of the rather opaque canvas -> imagick.Image -> Imagemagick.Image
  chain, although use of imagick is well within plugin parameters, by using it
  adhoc to generate external images and then integrated to a canvas in process)
- move start of imagmagick text functionality to separate plugin


### warhola 0.0.6 (07.06.2018)
- change /plugins/builtins to core
- additions to core functionality: adjust, blend, blur, convolute, noise, text,
  transform, translate 
- geo.Geometry lib for geometry abstraction across all commands
- started imagick integration 


### warhola 0.0.5 (18.05.2018)
- rewrite of canvas to eliminate bilevel canvas/image distortion and create 
  a single structure and interface for package needs (as well as compatibility 
  with image.Image)
- Operator interface for common operations externally available to mutate canvas
  internals


### warhola 0.0.4 (14.03.2018)
- cleanup & document 
- plugin abstraction to accomodate built in functionality
- added transform plugins: crop, resize, rotate, flip, shear, translate
- added adjustment plugins: brightness, gamma, contrast, hue, saturation 
- removed plugins to work on canvas issues


### warhola 0.0.3 (25.11.2017)
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


### warhola 0.0.2 (31.10.2017)
- function based Anchors


### warhola 0.0.1 (31.10.2017)
- initialize with changelog & readme 
