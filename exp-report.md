# Issues

[Issues][issues] is the worlds worst GitHub polling app. It polls
ksonnet/ksonnet for issues that are tagged "bug" and renders them into a graph.
This document recounts the process of hacking this together and trying to get it
to work using ksonnet.

## Proposals

At a high level, we need to reduce the amount of time it takes to get to one of
the "ah ha" moments. When you read the report you'll see they come a bit late.

Another important thing is that _composing_ components is very hard. Some of
this, which gets a more detailed treatment below, involves natively exposing
things like assuming service discovery exists (via `kubedns` or similar), as
well as things like a service being able to automatically pick up labels and
things.

There's a lot of other important but smaller things here, too, but the big one
we should think carefully about is the vscode extension, which could help
development, but currently does not.

* Transition prototypes to be closer to Jsonnet, so they're more familiar to
  people.
* It's nearly impossible to do anything without autocomplete working in vscode.
  We _have_ to fix this.
* kubedns- and env-based discovery should be natively supported in prototypes.
  * Every prototype should have these things bundled in by default.
  * We might need to be able to interrogate kubedns from `ks`. Possibly using an
    ephemeral container.
  * See the documentation about how to do this [here][dns], specifically the
    section titled "Accessing the Service".
* It would be very helpful for a service to be able to automatically pick up the
  labels of other objects, like a deployment.
  * Might be worth having `--labels` be a primitive in the same way that
    `--name` and `--namespace` are (or will be).
* Generate the prototype help messages using the same techniques we use to
  generate the README.
* Find a way to get the public IP of a service, if one exists?
* Debugging Jsonnet is hard. We should consider adding something like `printf`
  as a native function for Jsonnet.
* Create shorter aliases for commands?
* Scope prototypes at the level of registry and package. Then they don't have to
  be globally unique.
* Let's rename prototypes -> examples.
* Consistent help messages.
  * "This was your error. TRY: $thing"
  * All help message lines should probably wrap at 80 characters if we can help
    it.
* Consistent displaying of information in help messages.
* Make `--name` an implicit parameter for every prototype? I thought this was
  done already.
* The way namespaces are implicit behind the `apply` is a bit confusing. The
  proposal was to make namespaces explicit, right? We should probably make
  `--namespace` a flag that exists in every prototype, and it should default to
  the explicit namespace if not provided.
* Think hard about how we want to make the story easier for integrating with
  ingress. Not everything will have a `LoadBalancer` service.

### Smaller things

* Should be a nice offline experience.
  * `ks pkg describe` does not display prototypes correctly if package is not
    installed.
* Output of columnar data (_e.g._, `ks pkg list`) is not great.
* Maybe some commands should be runnable outside an app. (_e.g._, `prototype
  list`?)
* Consider making the default env name something other than `"default"` (can't
  always be the name of the context because context allows symbols we disallow).

## Journal

The following is exactly what I did, completely stream of consciousness:

* Wrote + Dockerized the world's worse GitHub polling app, that uses chart.js to
  render open issues marked "bug" in a graph. See [repository][issues].
* Run `ks pkg describe` does not correctly output the list of prototypes when the
* Run `ks pkg install incubator/redis`. Now when I run `ks pkg describe` I
  notice that it does list the prototypes. (I remember why it does this: we're
  trying to traverse the `prototypes` directory locally.)
* Try to run `ks prototype describe redis-stateless`. I think: man, that's more
  typing than I want. Maybe `describe` could present a menu of numbers and the
  user could just do `ks prototype describe 1` or something.
* I run `ks generate redis-stateless --name counts` and get an error: `ERROR
  Command is missing argument 'componentName'`. Weird, I thought `--name` was
  "required" now? Also, formatting for these help messages is just all over the
  place. We should use ANSI instead of `*` characters, for example.
* I give up and just give it the `componentName` argument.
* What's the default environment name again? `'default'`? Ugh, we should
  probably try to name that after the context. That could be hard, though,
  because contexts often have symbols that can't be in our names. Hmm.
* I run `ks show`. What's the default namespace I'm in, again? Namespace isn't
  in any of these resources. (Did the explicit namespace changes land? I guess
  they did.) I run `ks env list` and am greeted with this:

  ```
  NAME    NAMESPACE     SERVER
  ====    =========     ======
  default ks-playground https://heptio-ku-apiloadb-1ffzfwgtl8w34-1183314563.us-west-1.elb.amazonaws.com
  ```

  I honestly have no idea where `ks-playground` comes from. I run `kubectl
  config get-contexts` and realize I set this up some other time. Ugh. I guess
  now I have to delete that and start over.
* I run `ks help env` and `ks help env add` and skim until I figure out what to
  run: `ks env add dev-alex` and `ks env rm default`. Ok, well, at least it was
  easy. We're back in business.
* **REDIS IS RUNNING!** I think, this a-ha moment comes a bit too late.
* Ok, now I have several questions:
  * How do I get my app to interact with Redis?
  * How do I get my code into this app?
* `ks prototype list` again to see what's up. Oh, man, we really need to change
  prototypes to be hierarchical. This is a nightmare to figure out, as it is.
* Description messages are pretty good though, and I find the one I want pretty
  quick. I run `ks prototype describe deployed-service`. Man, that's a lot of
  typing.
* I run `ks generate deployed-service graphs --name issues --image
  hausdorff/issues:git-f2f9d15 --containerPort 8080 --type LoadBalancer`. There
  is a _lot_ of magic in that command. Prototype descriptions should have
  examples and stuff. There's no way that's going to work for most people.
* `ks apply dev-alex` again.
* How do I get the IP of a service again? `kubectl get svc issues -o yaml`.
  Gross! There's got to be a better way.
* I try to hit the endpoint. Hmm, it doesn't seem to be coming up. I wonder why.
  Maybe the LB is spinning up? This would _really_ suck with an ingress stack, I
  think?
* **IT SERVES THE PAGE!** Second a-ha moment is good, but also probably comes a
  bit too late.
* Ok, now I need my graphs to talk to redis. `ks pkg describe incubator/redis`
  is pretty pitiful. Which is sad, because the `README.md` in ksonnet/parts is
  so good. Can we re-use that? (Yes, we can, we wrote it especially to make that
  possible.)
* Ok so the readme doesn't tell me what I need to know. Let's look at the
  component. Is there a `ks component show`? Nope. Guess we have to crack open
  the editor. Man, the lack of strong vscode static analysis is just killing me.
  I can barely get anything done.
* Ah, looks like we're exposing a service as part of this prototype. Hmm.
* **KUBEDNS!** I run (guessing) `kubectl run curl
  --image=radial/busyboxplus:curl -i --tty`, and then `nslookup cache` (`cache`
  being the name of my redis deployment). This looks like it's going to work.
  Now we're really cooking.
* Ok, so it looks like I need to know what port to connect on, but that's no big
  deal. Problem solved, basically.

[issues]: https://github.com/hausdorff/issues
[dns]: https://kubernetes.io/docs/concepts/services-networking/connect-applications-service
