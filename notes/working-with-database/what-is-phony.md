{%hackmd BJrTq20hE %}
# What is Phony?
###### tags: `Makefile` `build tool`
## Phony is not the name of a file
A phony target is one that is not really the name of a file; rather it is just a name for a recipe to be executed when you make an explicit request. 

## Why use it?

There are two reasons to use a phony target
- to avoid a conflict with a file of the same name
- to improve performance.