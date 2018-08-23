module Graph exposing
    ( Config
    , Model
    , Msg(..)
    , new
    , subscriptions
    , update
    , view
    )

import Api
import Browser.Dom as Dom
import Browser.Events as Bev
import Html exposing (Html)
import Html.Attributes as Attr
import Http
import Process
import Task
import Url.Builder as Builder



-- CONFIG


type alias Config model msg =
    { get : model -> Model
    , set : model -> Model -> model
    , wrap : Msg -> msg
    }



-- MODEL


type alias Info =
    { metric : String
    , width : Int
    }


type State
    = Nothing
    | Delay String
    | Image Info


type Model
    = Model
        { state : State
        , width : Maybe Int
        , counter : Int
        }


new : Config model msg -> ( Model, Cmd msg )
new config =
    ( Model
        { state = Nothing
        , width = Maybe.Nothing
        , counter = 0
        }
    , Cmd.map config.wrap <|
        Task.perform viewportToMsg Dom.getViewport
    )



-- MSG


type Msg
    = Draw String
    | Resize Int Int
    | ResizeTimer Int


viewportToMsg : Dom.Viewport -> Msg
viewportToMsg { viewport } =
    Resize (truncate viewport.width) (truncate viewport.height)



-- SUBSCRIPTIONS


subscriptions : Config model msg -> model -> Sub msg
subscriptions config model =
    doSubscriptions (config.get model)
        |> Sub.map config.wrap


doSubscriptions : Model -> Sub Msg
doSubscriptions (Model model) =
    Bev.onResize Resize



-- UPDATE


update : Config model msg -> model -> Msg -> ( model, Cmd msg )
update config model msg =
    doUpdate config (config.get model) msg
        |> Tuple.mapFirst (config.set model)
        |> Tuple.mapSecond (Cmd.map config.wrap)


doUpdate : Config model msg -> Model -> Msg -> ( Model, Cmd Msg )
doUpdate config (Model model) msg =
    case msg of
        Draw metric ->
            ( case model.width of
                Maybe.Nothing ->
                    Model { model | state = Delay metric }

                Just width ->
                    Model { model | state = Image (Info metric width) }
            , Cmd.none
            )

        Resize width _ ->
            let
                newCounter =
                    model.counter + 1
            in
            ( Model { model | width = Just width, counter = newCounter }
            , sendDelayedMessage (ResizeTimer newCounter)
            )

        ResizeTimer counter ->
            ( case ( counter == model.counter, model.state, model.width ) of
                ( True, Delay metric, Just width ) ->
                    Model { model | state = Image (Info metric width) }

                ( True, Image info, Just width ) ->
                    Model { model | state = Image (Info info.metric width) }

                _ ->
                    Model model
            , Cmd.none
            )



-- VIEW


view : Config model msg -> Model -> Html msg
view config model =
    doView model
        |> Html.map config.wrap


doView : Model -> Html Msg
doView (Model model) =
    case model.state of
        Image { metric, width } ->
            let
                add key val params =
                    Builder.string key val :: params

                query =
                    []
                        |> add "metric" metric
                        |> add "width" (String.fromInt (width - 38))
                        |> add "padding" "5"
                        |> Builder.toQuery
            in
            Html.img [ Attr.src <| "/api/render" ++ query ] []

        _ ->
            Html.text ""



-- UTILS


sendMessage : msg -> Cmd msg
sendMessage msg =
    Task.perform (always msg) (Task.succeed ())


sendDelayedMessage : msg -> Cmd msg
sendDelayedMessage msg =
    Task.perform (always msg) (Process.sleep 250)
